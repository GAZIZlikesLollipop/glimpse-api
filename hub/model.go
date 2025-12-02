package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	connections map[int64]*websocket.Conn
	mutex       sync.RWMutex
}
