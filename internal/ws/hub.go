package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mer-prog/hookrelay-api/internal/model"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run processes register/unregister/broadcast events until the context is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			return
		case client := <-h.register:
			h.clients[client] = true
			slog.Info("ws client connected", "user_id", client.userID)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				slog.Info("ws client disconnected", "user_id", client.userID)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastDeliveryUpdate serialises a delivery log and sends it to all connected clients.
func (h *Hub) BroadcastDeliveryUpdate(log *model.DeliveryLog) {
	data, err := json.Marshal(log)
	if err != nil {
		slog.Error("ws broadcast marshal error", "error", err)
		return
	}
	h.broadcast <- data
}
