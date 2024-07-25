package connection

import (
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"io"
	connection_structs "sfu/handler/connection-structs"
	webrtcsfu "sfu/internal/webrtc-sfu"
	"time"
)

func instructorConnectionHandler(client *connection_structs.InstructorClient) error {
	fmt.Println("InstructorConnectionHandler")
	authPayload := make(chan connection_structs.InitialPayload)
	var msg connection_structs.InitialPayload
	var classId string
	//var classId string
	go func() {
		var payload connection_structs.InitialPayload
		time.Sleep(time.Second)
		err := client.Conn.ReadJSON(&payload)
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
		if msg.UserId == "" || classId == "" || connection_structs.Classes[msg.ClassId] == nil || connection_structs.Classes[msg.ClassId].IsLive == false {
			err := client.Conn.WriteJSON(&connection_structs.InitialSuccess{Success: false})
			if err != nil {
				fmt.Println(err.Error())
			}
			err = client.Conn.Close()
			if err != nil {
				fmt.Println(err.Error())

			}
			//return errors.New("class does not exist or the authentication of instructor failed, or class started")
		}

		//connection_structs.Classes[classId].Events =
		// do DB authentication here, do a simple match here

		fmt.Println("Authentication Done")
		err := client.Conn.WriteJSON(&connection_structs.InitialSuccess{Success: true})
		connection_structs.Classes[classId].IsLive = true
		connection_structs.Classes[classId].Instructor = &connection_structs.InstructorClient{
			Id:      msg.UserId,
			Conn:    client.Conn,
			ClassId: classId,
		}
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	case <-time.After(time.Second * 30):
		fmt.Println("Authentication Timed Out")
		err := client.Conn.WriteJSON(&connection_structs.InitialSuccess{Success: false})
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return errors.New("authentication Timed Out")
	}
	var peerConnection *webrtc.PeerConnection
	for {
		fmt.Println("Attempting to read from the instructor ", client.Conn.RemoteAddr())
		var event connection_structs.Event
		err := client.Conn.ReadJSON(&event)
		fmt.Println("UNMARSHALLED JSON", event.EventType)
		if err != nil {
			fmt.Println("ERROR IN LISTENING", err.Error())
			if err == io.EOF {
				fmt.Println("Connection Closed")
				connection_structs.Classes[classId].Instructor = nil
			}
			break
		}
		if event.EventType != connection_structs.ChatEventType && event.EventType != connection_structs.WebRTCEventType {
			connection_structs.Classes[classId].Events <- event
		} else if event.EventType == connection_structs.WebRTCEventType {
			peerConnection = webrtcsfu.InitializeSfuConnection(event, client.Conn, connection_structs.Instructor, classId)
		} else {
			connection_structs.Classes[classId].Chats <- connection_structs.Chat{
				EventType: event.EventType,
				Text:      event.Chat.Text,
				From:      event.Chat.From,
			}
		}
	}
	defer func() {
		if peerConnection == nil {
			return
		}
		cErr := peerConnection.Close()
		if cErr != nil {
			fmt.Printf("Error closing peer connection: %s", cErr.Error())
		}
	}()
	return nil
}
