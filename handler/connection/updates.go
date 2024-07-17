package connection

import "github.com/gorilla/websocket"

type UpdatesClient struct {
	conn *websocket.Conn
}
