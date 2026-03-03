package delivery

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mer-prog/hookrelay-api/internal/circuit"
	"github.com/mer-prog/hookrelay-api/internal/model"
	"github.com/mer-prog/hookrelay-api/internal/pubsub"
)

const defaultJobBuffer = 1000

// Engine manages a pool of workers that deliver webhooks.
type Engine struct {
	pool        *pgxpool.Pool
	pubsub      pubsub.PubSub
	breakerMgr  *circuit.BreakerManager
	workerCount int
	jobChan     chan DeliveryJob
	cancel      context.CancelFunc
}

func NewEngine(pool *pgxpool.Pool, ps pubsub.PubSub, bm *circuit.BreakerManager, workerCount int) *Engine {
	return &Engine{
		pool:        pool,
		pubsub:      ps,
		breakerMgr:  bm,
		workerCount: workerCount,
		jobChan:     make(chan DeliveryJob, defaultJobBuffer),
	}
}

// Start launches the worker goroutines.
func (e *Engine) Start(ctx context.Context) {
	ctx, e.cancel = context.WithCancel(ctx)
	for i := 0; i < e.workerCount; i++ {
		w := newWorker(i, e.jobChan, e.pool, e.pubsub, e.breakerMgr, 30000)
		go w.Start(ctx)
	}
	slog.Info("delivery engine started", "workers", e.workerCount)
}

// Dispatch sends delivery jobs for each matching endpoint.
func (e *Engine) Dispatch(ctx context.Context, event model.Event, endpoints []model.Endpoint) {
	for _, ep := range endpoints {
		select {
		case e.jobChan <- DeliveryJob{Event: event, Endpoint: ep}:
		case <-ctx.Done():
			return
		}
	}
}

// Stop shuts down all workers and closes the job channel.
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
	close(e.jobChan)
	slog.Info("delivery engine stopped")
}
