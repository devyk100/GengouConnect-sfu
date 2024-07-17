package connection

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sfu/internal"
)

type InitialSuccess struct {
	Success bool `json:"success"`
}

type InitialPayload struct {
	UserId  string `json:"userId"`
	ClassId string `json:"classId"`
}

const (
	Instructor = "0"
	Learner    = "1"
	Updates    = "2"
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

	if request.URL.Query().Get("type") == Instructor {
		client := &InstructorClient{conn: conn}
		err := instructorConnectionHandler(client)
		if err != nil {
			fmt.Println(err.Error(), "In making the instructor handler")
			return
		}
		//func(client *InstructorClient) {
		//
		//}(client)
		return
	} else if request.URL.Query().Get("type") == Learner {
		client := &LearnerClient{conn: conn}
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
