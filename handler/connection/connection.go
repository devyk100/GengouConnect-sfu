package connection

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	connectionData "sfu/handler/connection-structs"
	"sfu/internal"
)

func HandleInitConnection(writer http.ResponseWriter, request *http.Request) {
	/*
		Upgrading this connection to WS at last using gorilla websockets
	*/
	conn, err := internal.CrossOrigin.Upgrade(writer, request, nil)

	/*
		At exit of this function close this websocket connection.
	*/
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}(conn)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	/*
		based on the URL parameter of the websocket connection we connection them as a learner or as an instructor. And they are stalling functions
	*/
	if request.URL.Query().Get("type") == connectionData.Instructor {

		client := &connectionData.InstructorClient{Conn: conn}

		err := instructorConnectionHandler(client)

		if err != nil {
			fmt.Println(err.Error(), "In making the instructor handler")
			return
		}

		return

	} else if request.URL.Query().Get("type") == connectionData.Learner {
		client := &connectionData.LearnerClient{Conn: conn}
		err := learnerConnectionHandler(client)
		if err != nil {
			fmt.Println(err.Error(), "In making the learner handler")
			return
		}
		return
	} else {
		// MAKE THE UPDATE HANDLER
	}

}
