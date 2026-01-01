package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/timmaaaz/ichor/foundation/logger"
)

const (
	// writeWait is the maximum time allowed to write a message to the peer.
	// 10 seconds is generous for most networks; increase if high-latency clients expected.
	writeWait = 10 * time.Second

	// pongWait is the maximum time to wait for a pong response from the peer.
	// 60 seconds allows for occasional network hiccups without disconnecting.
	pongWait = 60 * time.Second

	// pingPeriod determines how often to send ping messages to the peer.
	// Set to 90% of pongWait to ensure we send a ping before the pong deadline.
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize is the maximum size of incoming messages from the peer.
	// 512 bytes is sufficient for client commands (ack, ping). Alerts are server→client only.
	maxMessageSize = 512

	// sendBufferSize is the capacity of the outgoing message buffer.
	// 256 messages handles burst scenarios. If consistently full, consider back-pressure.
	sendBufferSize = 256
)

// Client represents a WebSocket connection.
// It is generic and has no business logic - application-specific semantics
// are added by higher layers (e.g., AlertHub).
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	ids  []string     // IDs this client is registered under
	idMu sync.RWMutex // Protects ids for concurrent access
	send chan []byte
	done chan struct{}
	log  *logger.Logger
}

// NewClient creates a new Client.
func NewClient(hub *Hub, conn *websocket.Conn, log *logger.Logger) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, sendBufferSize),
		done: make(chan struct{}),
		log:  log,
	}
}

// IDs returns a copy of the client's registered IDs.
// Thread-safe.
func (c *Client) IDs() []string {
	c.idMu.RLock()
	defer c.idMu.RUnlock()
	return append([]string(nil), c.ids...)
}

// SetIDs updates the client's registered IDs.
// Thread-safe.
func (c *Client) SetIDs(ids []string) {
	c.idMu.Lock()
	defer c.idMu.Unlock()
	c.ids = append([]string(nil), ids...)
}

// Send queues a message for sending to the client.
// Non-blocking; drops message if buffer is full.
func (c *Client) Send(ctx context.Context, message []byte) {
	select {
	case c.send <- message:
	case <-c.done:
		// Client is closing, skip
	default:
		// Buffer full - log and drop message
		c.log.Warn(ctx, "websocket send buffer full, message dropped")
	}
}

// Close signals the client to stop and closes the connection.
// Safe to call multiple times.
func (c *Client) Close() error {
	select {
	case <-c.done:
		return nil // Already closing
	default:
		close(c.done)
	}
	return nil
}

// ReadPump pumps messages from the WebSocket connection.
// Blocks until the connection is closed.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(ctx, c)
		c.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)

	for {
		_, _, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				c.log.Info(ctx, "websocket connection closed normally")
			} else {
				c.log.Info(ctx, "websocket read error", "error", err)
			}
			break
		}
		// For now, we don't process incoming messages from client
		// Future: handle acknowledgments, etc.
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
// Owns the connection lifecycle - responsible for closing the underlying connection.
// Note: coder/websocket's Close() is safe to call multiple times and will return
// an error if already closed, which we ignore in the defer.
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// Close is safe to call even if ReadPump already closed the connection.
		// coder/websocket handles this gracefully.
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}

			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Write(writeCtx, websocket.MessageText, message)
			cancel()

			if err != nil {
				c.log.Warn(ctx, "websocket write error", "error", err)
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Ping(pingCtx)
			cancel()

			if err != nil {
				c.log.Warn(ctx, "websocket ping error", "error", err)
				return
			}

		case <-c.done:
			return

		case <-ctx.Done():
			return
		}
	}
}
