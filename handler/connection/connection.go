package connection

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	connection_structs "sfu/handler/connection-structs"
	"sfu/internal"
)

func HandleInitConnection(writer http.ResponseWriter, request *http.Request) {
	conn, err := internal.CrossOrigin.Upgrade(writer, request, nil)
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

	if request.URL.Query().Get("type") == connection_structs.Instructor {
		client := &connection_structs.InstructorClient{Conn: conn}
		err := instructorConnectionHandler(client)
		if err != nil {
			fmt.Println(err.Error(), "In making the instructor handler")
			return
		}
		//func(client *InstructorClient) {
		//
		//}(client)
		return
	} else if request.URL.Query().Get("type") == connection_structs.Learner {
		client := &connection_structs.LearnerClient{Conn: conn}
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
