package websocket

import (
	"sync"
	"github.com/gorilla/websocket"
)

// Client representa una conexión activa de un dispositivo
type Client struct {
	UserID string
	Conn   *websocket.Conn
}

// Hub gestiona todas las conexiones activas
type Hub struct {
	// Mapeamos UserID -> Lista de conexiones (dispositivos)
	clients map[string][]*websocket.Conn
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string][]*websocket.Conn),
	}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = append(h.clients[userID], conn)
}

func (h *Hub) Unregister(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.clients[userID]
	for i, c := range conns {
		if c == conn {
			h.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

// NotifyUser envía una señal de "sync_needed" a todos los dispositivos del usuario
func (h *Hub) NotifyUser(userID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if devices, ok := h.clients[userID]; ok {
		for _, conn := range devices {
			// Enviamos un mensaje ligero
			conn.WriteJSON(map[string]string{"type": "sync_needed"})
		}
	}
}
