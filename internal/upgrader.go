package internal

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var CrossOrigin = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
