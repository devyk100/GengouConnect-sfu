package connection

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"time"
)

type InstructorClient struct {
	id      string
	classId string
	conn    *websocket.Conn
}

func instructorConnectionHandler(client *InstructorClient) error {
	fmt.Println("InstructorConnectionHandler")
	authPayload := make(chan InitialPayload)
	var msg InitialPayload
	var classId string
	//var classId string
	go func() {
		var payload InitialPayload
		time.Sleep(time.Second)
		err := client.conn.ReadJSON(&payload)
		fmt.Println(payload.ClassId, payload.UserId)
		authPayload <- payload
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()

	select {
	case msg = <-authPayload:

		// VERIFY IF THE INSTRUCTOR FROM THE INSTRUCTOR ID IS ALLOWED ON THIS CLASS ID FROM THE DATABASE.
		classId = msg.ClassId
		if msg.UserId == "" || classId == "" || Classes[msg.ClassId] == nil || Classes[msg.ClassId].isLive == false {
			err := client.conn.WriteJSON(&InitialSuccess{Success: false})
			if err != nil {
				fmt.Println(err.Error())
			}
			err = client.conn.Close()
			if err != nil {
				fmt.Println(err.Error())

			}
			//return errors.New("class does not exist or the authentication of instructor failed, or class started")
		}

		//Classes[classId].Events =
		// do DB authentication here, do a simple match here

		fmt.Println("Authentication Done")
		err := client.conn.WriteJSON(&InitialSuccess{Success: true})
		Classes[classId].isLive = true
		Classes[classId].Instructor = &InstructorClient{
			id:      msg.UserId,
			conn:    client.conn,
			classId: classId,
		}
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	case <-time.After(time.Second * 30):
		fmt.Println("Authentication Timed Out")
		err := client.conn.WriteJSON(&InitialSuccess{Success: false})
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return errors.New("authentication Timed Out")
	}
	for {
		fmt.Println("Attempting to read from the instructor ", client.conn.RemoteAddr())
		var event BoardEvent
		err := client.conn.ReadJSON(&event)
		fmt.Println("UNMARSHALLED JSON", event.EventType)
		if err != nil {
			fmt.Println("ERROR IN LISTENING", err.Error())
			if err == io.EOF {
				fmt.Println("Connection Closed")
				Classes[classId].Instructor = nil
			}
			break
		}
		if event.EventType != ChatEventType {
			Classes[classId].Events <- event
		} else {
			Classes[classId].Chats <- Chat{
				EventType: event.EventType,
				Text:      event.Chat.Text,
				From:      event.Chat.From,
			}
		}
	}
	return nil
}
