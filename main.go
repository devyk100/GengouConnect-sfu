package main

import (
	"fmt"
	"net/http"
	"sfu/handler/connection"
)

func main() {
	/*
		This is the main websocket connection
	*/
	http.HandleFunc("/ws", connection.HandleInitConnection)

	/*
		This is the class creation
	*/
	http.HandleFunc("/create-class", connection.CreateClassHandler)

	/*
		This is for the recent chats of the class, an HTTP handler
	*/
	http.HandleFunc("/chats", connection.PreviousChatHandler)
	fmt.Println("Listening on port 8000")

	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		panic(err.Error())
		return
	}
}
