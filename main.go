package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sfu/handler/connection"
)

type Class struct {
	TeacherConn *websocket.Conn
	StudentConn []*websocket.Conn
}

type Client struct {
	conn *websocket.Conn
}

type Message struct {
	Message string `json:"message"`
}

var clients = make(map[*Client]bool)

var broadCastedMessage = make(chan Message)

func handleConnections(writer http.ResponseWriter, request *http.Request) {
	conn := &websocket.Conn{}
	var err error
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			panic(err.Error())
		}
	}(conn)

	if err != nil {
		panic(err.Error())
	}
	client := &Client{conn: conn}
	clients[client] = true

	for {
		fmt.Println("Attempting to read from ", client.conn.RemoteAddr())
		var msg Message
		err := client.conn.ReadJSON(&msg)
		fmt.Println("Got message: ", msg.Message)
		if err != nil {
			fmt.Println("Error reading from", client.conn.RemoteAddr(), ". Disconnecting")
			break
		}
		broadCastedMessage <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadCastedMessage
		for client := range clients {
			fmt.Println("SENDING MESSAGE TO ", client.conn.RemoteAddr())
			err := client.conn.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				err := client.conn.Close()
				if err != nil {
					return
				}
				return
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", connection.HandleInitConnection)
	http.HandleFunc("/create-class", connection.CreateClassHandler)
	http.HandleFunc("/chats", connection.PreviousChatHandler)
	go handleMessages()
	fmt.Println("Listening on port 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err.Error())
		return
	}
}
