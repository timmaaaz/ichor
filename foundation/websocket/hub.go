package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/timmaaaz/ichor/foundation/logger"
)

// Hub manages WebSocket connections and message broadcasting.
// Uses string-based IDs for maximum flexibility - the calling application
// defines the semantics (e.g., "user:{uuid}", "role:{uuid}").
type Hub struct {
	// clients maps ID to the set of clients registered under that ID.
	// A single client can be registered under multiple IDs.
	clients map[string]map[*Client]bool

	// clientIDs tracks which IDs each client is registered under.
	// Used for efficient unregistration and ID updates.
	clientIDs map[*Client][]string

	mu  sync.RWMutex
	log *logger.Logger
}

// NewHub creates a new Hub instance.
func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:   make(map[string]map[*Client]bool),
		clientIDs: make(map[*Client][]string),
		log:       log,
	}
}

// Run starts the hub and blocks until context is cancelled.
// Periodically logs connection metrics for monitoring.
func (h *Hub) Run(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.mu.RLock()
			connCount := len(h.clientIDs)
			idCount := len(h.clients)
			h.mu.RUnlock()

			h.log.Debug(ctx, "websocket metrics",
				"total_connections", connCount,
				"unique_ids", idCount)
		case <-ctx.Done():
			h.log.Info(ctx, "websocket hub shutting down")
			return ctx.Err()
		}
	}
}

// Register adds a client to the hub under the given IDs.
func (h *Hub) Register(ctx context.Context, client *Client, ids []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Store the IDs for this client
	h.clientIDs[client] = ids
	client.SetIDs(ids)

	// Register under each ID
	for _, id := range ids {
		if h.clients[id] == nil {
			h.clients[id] = make(map[*Client]bool)
		}
		h.clients[id][client] = true
	}

	h.log.Info(ctx, "websocket client registered",
		"id_count", len(ids),
		"total_connections", len(h.clientIDs))
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(ctx context.Context, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ids, ok := h.clientIDs[client]
	if !ok {
		return // Client not registered
	}

	// Remove from each ID map
	for _, id := range ids {
		if clients, ok := h.clients[id]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.clients, id)
			}
		}
	}

	delete(h.clientIDs, client)

	h.log.Info(ctx, "websocket client unregistered",
		"total_connections", len(h.clientIDs))
}

// BroadcastToID sends a message to all connections registered under the given ID.
// Uses context.Background() intentionally - broadcasts should not be cancelled by
// the caller's context since they're fire-and-forget operations. Individual client
// Send() calls handle their own timeouts via the write pump.
func (h *Hub) BroadcastToID(id string, message []byte) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.clients[id]
	for client := range clients {
		client.Send(context.Background(), message)
	}
	return len(clients)
}

// BroadcastAll sends a message to all connected clients.
func (h *Hub) BroadcastAll(message []byte) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Use clientIDs to iterate unique clients (avoid duplicates)
	for client := range h.clientIDs {
		client.Send(context.Background(), message)
	}
	return len(h.clientIDs)
}

// UpdateClientIDs updates the IDs for a specific client.
func (h *Hub) UpdateClientIDs(ctx context.Context, client *Client, newIDs []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	oldIDs := h.clientIDs[client]

	// Remove from old ID maps
	for _, id := range oldIDs {
		if clients, ok := h.clients[id]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.clients, id)
			}
		}
	}

	// Add to new ID maps
	h.clientIDs[client] = newIDs
	client.SetIDs(newIDs)
	for _, id := range newIDs {
		if h.clients[id] == nil {
			h.clients[id] = make(map[*Client]bool)
		}
		h.clients[id][client] = true
	}
}

// UpdateClientIDsForID updates IDs for all clients registered under the given ID.
// Used for bulk updates (e.g., update all clients for a user when roles change).
// Note: This matches the exact ID, not a prefix pattern. For user role updates,
// pass the full user ID string (e.g., "user:uuid") to update all that user's connections.
func (h *Hub) UpdateClientIDsForID(ctx context.Context, id string, newIDs []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Collect clients to update (we need to iterate before modifying)
	clientsToUpdate := make([]*Client, 0)
	for client := range h.clients[id] {
		clientsToUpdate = append(clientsToUpdate, client)
	}

	for _, client := range clientsToUpdate {
		oldIDs := h.clientIDs[client]

		// Remove from old ID maps
		for _, oldID := range oldIDs {
			if clientSet, ok := h.clients[oldID]; ok {
				delete(clientSet, client)
				if len(clientSet) == 0 {
					delete(h.clients, oldID)
				}
			}
		}

		// Add to new ID maps
		h.clientIDs[client] = newIDs
		client.SetIDs(newIDs)
		for _, newID := range newIDs {
			if h.clients[newID] == nil {
				h.clients[newID] = make(map[*Client]bool)
			}
			h.clients[newID][client] = true
		}
	}
}

// CloseAll closes all connected clients. Used for graceful shutdown.
func (h *Hub) CloseAll(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clientIDs {
		client.Close()
	}

	h.log.Info(ctx, "closed all websocket connections",
		"count", len(h.clientIDs))
	return nil
}

// ConnectionCount returns the total number of active connections.
func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clientIDs)
}

// ClientsForID returns the number of clients registered under the given ID.
func (h *Hub) ClientsForID(id string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[id])
}
