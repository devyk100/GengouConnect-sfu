package connection

import (
	"errors"
	"fmt"
	"github.com/pion/webrtc/v4"
	"io"
	connectionstructs "sfu/handler/connection-structs"
	webrtcsfu "sfu/internal/webrtc-sfu"
	"time"
)

func learnerConnectionHandler(client *connectionstructs.LearnerClient) error {
	fmt.Println("Learner Client Connected")
	authPayload := make(chan connectionstructs.InitialPayload)
	var msg connectionstructs.InitialPayload
	var classId string

	go func() {
		var payload connectionstructs.InitialPayload
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

		// VERIFY IF THE LEARNER FROM THE LEARNER ID IS ALLOWED ON THIS CLASS ID FROM THE DATABASE.
		classId = msg.ClassId
		if msg.UserId == "" || classId == "" || connectionstructs.Classes[msg.ClassId] == nil || connectionstructs.Classes[msg.ClassId].IsLive == false {
			fmt.Println("DEBUG: UserId", msg.UserId, "ClassId", classId, "Does classes exist")
			err := client.Conn.WriteJSON(&connectionstructs.InitialSuccess{Success: false})
			if err != nil {
				fmt.Println(err.Error())
			}
			err = client.Conn.Close()
			if err != nil {
				fmt.Println(err.Error())

			}
			return errors.New("class does not exist or the authentication of instructor failed, or class started")
		}
		if connectionstructs.Classes[classId].Learners == nil {
			connectionstructs.Classes[classId].Learners = []*connectionstructs.LearnerClient{}
		}
		connectionstructs.Classes[classId].LearnersLock.Lock()
		connectionstructs.Classes[classId].Learners = append(connectionstructs.Classes[classId].Learners, &connectionstructs.LearnerClient{
			Id:      msg.UserId,
			ClassId: classId,
			Conn:    client.Conn,
		})
		connectionstructs.Classes[classId].LearnersLock.Unlock()

		fmt.Println("Authentication Done")
		err := client.Conn.WriteJSON(&connectionstructs.InitialSuccess{Success: true})
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	case <-time.After(time.Second * 30):
		fmt.Println("Authentication Timed Out")
		err := client.Conn.WriteJSON(&connectionstructs.InitialSuccess{Success: false})
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return errors.New("authentication Timed Out")
	}
	var peerConnection *webrtc.PeerConnection
	for {
		// handling the chat
		payload := connectionstructs.Chat{}
		err := client.Conn.ReadJSON(&payload)
		if payload.EventType == connectionstructs.WebRTCEventType {
			peerConnection = webrtcsfu.InitializeSfuConnection(connectionstructs.Event{
				EventType:   payload.EventType,
				WebrtcEvent: payload.WebrtcEvent,
			}, client.Conn, connectionstructs.Learner, classId)
		}
		fmt.Println("Read from client")
		if err != nil {
			fmt.Println(err.Error())
			if err == io.EOF {
				fmt.Println("Connection Closed")
				connectionstructs.Classes[classId].Instructor = nil
			}
			break
		}
		connectionstructs.Classes[classId].Chats <- payload
	}
	defer func() {
		if peerConnection != nil {
			return
		}
		cErr := peerConnection.Close()
		if cErr != nil {
			fmt.Printf("Error closing peer connection: %s", cErr.Error())
		}
	}()
	return nil
}
