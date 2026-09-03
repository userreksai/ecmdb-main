package web

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/userreksai/ecmdb-main/pkg/term/guacx"
)

type Session struct {
	Websocket *websocket.Conn
	Tunnel    *guacx.Tunnel
	mutex     sync.Mutex
}

func (s *Session) Close() {
	if s.Tunnel != nil {
		_ = s.Tunnel.Close()
	}

	if s.Websocket != nil {
		_ = s.Websocket.Close()
	}
}
