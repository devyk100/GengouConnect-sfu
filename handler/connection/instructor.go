package connection

import (
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"io"
	connectionData "sfu/handler/connection-structs"
	webrtcsfu "sfu/internal/webrtc-sfu"
	"time"
)

func instructorConnectionHandler(client *connectionData.InstructorClient) error {
	fmt.Println("InstructorConnectionHandler")
	authPayload := make(chan connectionData.InitialPayload) // the initial payload channel to receive it from the initial websocket connection to do auth
	var msg connectionData.InitialPayload                   // the initial payload variable
	var classId string                                      // the class ID for the class where the instructor would be connecting

	go func() {
		var payload connectionData.InitialPayload
		err := client.Conn.ReadJSON(&payload) // we wait for the initial payload, which will be used inside auth
		fmt.Println(payload.ClassId, payload.UserId)
		authPayload <- payload // pass it to the channel for processing it in select clause
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()

	select {
	case msg = <-authPayload: // we get the payload in the msg variable

		// VERIFY IF THE INSTRUCTOR FROM THE INSTRUCTOR ID IS ALLOWED ON THIS CLASS ID FROM THE DATABASE.

		classId = msg.ClassId

		/*
			If the userid is null, class id is null, or the class wasn't started before hand, or the isLive variable was not set to true, in such case we return false and dump the websocket connection.
		*/
		if msg.UserId == "" || classId == "" || connectionData.Classes[msg.ClassId] == nil || connectionData.Classes[msg.ClassId].IsLive == false {
			err := client.Conn.WriteJSON(&connectionData.InitialSuccess{Success: false})
			if err != nil {
				fmt.Println(err.Error())
			}
			err = client.Conn.Close()
			if err != nil {
				fmt.Println(err.Error())
			}
			return errors.New("class does not exist or the authentication of instructor failed, or class started")
		}

		//connectionData.Classes[classId].Events =
		// do DB authentication here, do a simple match here

		/*
			Auth was success, so return the true message.
		*/
		fmt.Println("Authentication Done")
		err := client.Conn.WriteJSON(&connectionData.InitialSuccess{Success: true})
		connectionData.Classes[classId].Instructor = &connectionData.InstructorClient{
			Id:      msg.UserId,
			Conn:    client.Conn,
			ClassId: classId,
		}
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

	/*
		If after a certain timeout since when the connection was established, if we do not get auth credentials, we drop the connection.
	*/
	case <-time.After(time.Second * 30):
		fmt.Println("Authentication Timed Out")
		err := client.Conn.WriteJSON(&connectionData.InitialSuccess{Success: false})
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return errors.New("authentication Timed Out")
	}

	/*
		We continue with the rest of the connection here.
	*/
	var peerConnection *webrtc.PeerConnection
	/*
			We listen for:
			1. Other events including the board event
			2. WebRTC connection events.
			3. Chat events.
		Below
	*/
	for {
		fmt.Println("Attempting to read from the instructor ", client.Conn.RemoteAddr())

		var event connectionData.Event

		err := client.Conn.ReadJSON(&event)
		fmt.Println("UNMARSHALLED JSON", event.EventType)
		if err != nil {
			fmt.Println("ERROR IN LISTENING", err.Error())
			if err == io.EOF {
				fmt.Println("Connection Closed")
				connectionData.Classes[classId].Instructor = nil
			}
			break
		}

		if event.EventType != connectionData.ChatEventType && event.EventType != connectionData.WebRTCEventType {
			fmt.Println("We got a BOARD EVENT YAY!", event.BoardSvg)
			connectionData.Classes[classId].Events <- event
		} else if event.EventType == connectionData.WebRTCEventType {
			peerConnection = webrtcsfu.InitializeSfuConnection(event, client.Conn, connectionData.Instructor, classId)
		} else {
			connectionData.Classes[classId].Chats <- connectionData.Chat{
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
