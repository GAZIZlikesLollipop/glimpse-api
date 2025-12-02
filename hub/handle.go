package hub

import "github.com/gorilla/websocket"

func InitHub() *Hub {
	return &Hub{
		connections: make(map[int64]*websocket.Conn),
	}
}

func (h *Hub) AddConnection(userId int64, cnn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.connections[userId] = cnn
}

func (h *Hub) RemoveConnection(userId int64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.connections, userId)
}

func (h *Hub) GetConnections() map[int64]*websocket.Conn {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	result := make(map[int64]*websocket.Conn)
	for k, v := range h.connections {
		result[k] = v
	}
	return result
}

func (h *Hub) GetConnection(userId int64) (*websocket.Conn, bool) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	cnn, ok := h.connections[userId]
	return cnn, ok
}
