package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/circuit"
	"github.com/mer-prog/hookrelay-api/internal/model"
	"github.com/mer-prog/hookrelay-api/internal/pubsub"
)

// DeliveryJob represents a single webhook delivery task.
type DeliveryJob struct {
	Event    model.Event
	Endpoint model.Endpoint
}

// Worker pulls jobs from a channel and executes webhook deliveries.
type Worker struct {
	ID         int
	jobChan    <-chan DeliveryJob
	pool       *pgxpool.Pool
	pubsub     pubsub.PubSub
	breakerMgr *circuit.BreakerManager
	httpClient *http.Client
}

func newWorker(id int, jobChan <-chan DeliveryJob, pool *pgxpool.Pool, ps pubsub.PubSub, bm *circuit.BreakerManager, timeoutMs int) *Worker {
	return &Worker{
		ID:         id,
		jobChan:    jobChan,
		pool:       pool,
		pubsub:     ps,
		breakerMgr: bm,
		httpClient: &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}
}

// Start runs the worker loop until the context is cancelled.
func (w *Worker) Start(ctx context.Context) {
	slog.Info("worker started", "worker_id", w.ID)
	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopped", "worker_id", w.ID)
			return
		case job, ok := <-w.jobChan:
			if !ok {
				return
			}
			w.executeDelivery(ctx, job)
		}
	}
}

func (w *Worker) executeDelivery(ctx context.Context, job DeliveryJob) {
	cb := w.breakerMgr.GetBreaker(job.Endpoint.ID)
	if cb.IsOpen() {
		slog.Warn("circuit open, skipping delivery",
			"endpoint_id", job.Endpoint.ID,
			"event_id", job.Event.ID,
		)
		return
	}

	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	signature := Sign(job.Event.Payload, job.Endpoint.Secret, timestamp)

	reqHeaders := map[string]string{
		"Content-Type":        "application/json",
		"X-HookRelay-Event":  job.Event.EventType,
		"X-HookRelay-Sig":    signature,
		"X-HookRelay-Ts":     timestamp,
		"X-HookRelay-ID":     job.Event.ID.String(),
	}
	reqHeadersJSON, _ := json.Marshal(reqHeaders)

	dl := &model.DeliveryLog{
		EventID:        job.Event.ID,
		EndpointID:     job.Endpoint.ID,
		Status:         "PENDING",
		AttemptNumber:  1,
		RequestHeaders: reqHeadersJSON,
	}
	if err := model.CreateDeliveryLog(ctx, w.pool, dl); err != nil {
		slog.Error("failed to create delivery log", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.Endpoint.URL, bytes.NewReader(job.Event.Payload))
	if err != nil {
		w.recordFailure(ctx, dl, cb, nil, fmt.Sprintf("building request: %v", err), job.Endpoint.MaxRetries)
		return
	}
	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := w.httpClient.Do(req)
	latency := int(time.Since(start).Milliseconds())
	dl.LatencyMs = &latency

	if err != nil {
		w.recordFailure(ctx, dl, cb, nil, fmt.Sprintf("http request: %v", err), job.Endpoint.MaxRetries)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	bodyStr := string(body)
	dl.ResponseStatus = &resp.StatusCode
	dl.ResponseBody = &bodyStr
	respHeadersJSON, _ := json.Marshal(resp.Header)
	dl.ResponseHeaders = respHeadersJSON

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		dl.Status = "SUCCESS"
		dl.NextRetryAt = nil
		cb.RecordSuccess()
		_ = model.UpdateDeliveryLog(ctx, w.pool, dl)
	} else {
		errMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		w.recordFailure(ctx, dl, cb, &resp.StatusCode, errMsg, job.Endpoint.MaxRetries)
	}

	w.publishResult(ctx, dl)
}

func (w *Worker) recordFailure(ctx context.Context, dl *model.DeliveryLog, cb *circuit.CircuitBreaker, statusCode *int, errMsg string, maxRetries int) {
	dl.ErrorMessage = &errMsg
	cb.RecordFailure()

	if statusCode != nil && !ShouldRetry(*statusCode) {
		dl.Status = "FAILED"
		_ = model.UpdateDeliveryLog(ctx, w.pool, dl)
		return
	}

	if dl.AttemptNumber >= maxRetries {
		dl.Status = "FAILED"
		_ = model.UpdateDeliveryLog(ctx, w.pool, dl)
		return
	}

	if err := ScheduleRetry(ctx, w.pool, dl); err != nil {
		slog.Error("failed to schedule retry", "error", err)
	}
}

func (w *Worker) publishResult(ctx context.Context, dl *model.DeliveryLog) {
	data, err := json.Marshal(dl)
	if err != nil {
		return
	}
	_ = w.pubsub.Publish(ctx, "deliveries", data)
}
