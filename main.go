package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sfu/handler/connection"
)

type Client struct {
	conn *websocket.Conn
}

type Message struct {
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/ws", connection.HandleInitConnection)
	http.HandleFunc("/create-class", connection.CreateClassHandler)
	http.HandleFunc("/chats", connection.PreviousChatHandler)
	fmt.Println("Listening on port 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err.Error())
		return
	}
}
