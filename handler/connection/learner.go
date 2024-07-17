package connection

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"time"
)

type LearnerClient struct {
	id      string
	classId string
	conn    *websocket.Conn
}

func learnerConnectionHandler(client *LearnerClient) error {
	fmt.Println("Learner Client Connected")
	authPayload := make(chan InitialPayload)
	var msg InitialPayload
	var classId string

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

		// VERIFY IF THE LEARNER FROM THE LEARNER ID IS ALLOWED ON THIS CLASS ID FROM THE DATABASE.
		classId = msg.ClassId
		if msg.UserId == "" || classId == "" || Classes[msg.ClassId] == nil || Classes[msg.ClassId].isLive == false {
			fmt.Println("DEBUG: UserId", msg.UserId, "ClassId", classId, "Does classes exist")
			err := client.conn.WriteJSON(&InitialSuccess{Success: false})
			if err != nil {
				fmt.Println(err.Error())
			}
			err = client.conn.Close()
			if err != nil {
				fmt.Println(err.Error())

			}
			return errors.New("class does not exist or the authentication of instructor failed, or class started")
		}
		if Classes[classId].Learners == nil {
			Classes[classId].Learners = []*LearnerClient{}
		}
		Classes[classId].LearnersLock.Lock()
		Classes[classId].Learners = append(Classes[classId].Learners, &LearnerClient{
			id:      msg.UserId,
			classId: classId,
			conn:    client.conn,
		})
		Classes[classId].LearnersLock.Unlock()

		fmt.Println("Authentication Done")
		err := client.conn.WriteJSON(&InitialSuccess{Success: true})
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
		// handling the chat
		payload := Chat{}
		err := client.conn.ReadJSON(&payload)
		if err != nil {
			fmt.Println(err.Error())
			if err == io.EOF {
				fmt.Println("Connection Closed")
				Classes[classId].Instructor = nil
			}
			break
		}
		Classes[classId].Chats <- payload
	}

	return nil
}
