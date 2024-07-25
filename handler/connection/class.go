package connection

//yash@Om#12$
//yash@135
//ommehta@123

import (
	"fmt"
	"net/http"
	connection_structs "sfu/handler/connection-structs"
	webrtc_sfu "sfu/internal/webrtc-sfu"
	webrtc_turn "sfu/internal/webrtc-turn"
	"sync"
	"time"
)

func CreateClassHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("CreateClassHandler Called")
	classId := request.URL.Query().Get("classId")

	// Make a check if this was put in the Database or not by the backend

	writer.Header().Set("Content-Type", "application/json")
	if classId == "" || connection_structs.Classes[classId] != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, err := writer.Write([]byte(`{"success": false }`))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		return
	}
	if connection_structs.Classes[classId] == nil {
		connection_structs.Classes[classId] = &connection_structs.Class{
			Instructor:    nil,
			Learners:      nil,
			IsLive:        true,
			Events:        make(chan connection_structs.Event),
			Chats:         make(chan connection_structs.Chat, 50),
			ClassId:       classId,
			ChatsLock:     sync.RWMutex{},
			PreviousChats: make([]connection_structs.Chat, connection_structs.PreviousChatAmount),
			LearnersLock:  sync.RWMutex{},
		}

	}
	if webrtc_sfu.LiveClasses[classId] == nil {
		webrtc_sfu.LiveClasses[classId] = &webrtc_sfu.LiveClass{
			InstructorPeerConnection: nil,
			LearnerPeerConnections:   nil,
			ClassId:                  classId,
			LocalTrack:               nil,
		}
	}
	writer.WriteHeader(http.StatusCreated)

	go func(classId string) {
		for {
			select {
			case <-time.After(time.Minute * 60):
				if connection_structs.Classes[classId].Instructor == nil {
					fmt.Println("Delete the class room", classId)
					close(connection_structs.Classes[classId].Events)
					close(connection_structs.Classes[classId].Chats)
					connection_structs.Classes[classId] = nil
					return
				}
			}
		}
	}(classId)

	go connection_structs.Classes[classId].BoardEventHandler()
	go connection_structs.Classes[classId].ChatHandler()

	if !webrtc_turn.IsTurnStarted() {
		turnStopChan := webrtc_turn.TurnStarter()
		go func() {
			for {
				time.Sleep(time.Minute * 20)
				if len(webrtc_sfu.LiveClasses) < 1 {
					turnStopChan <- true
					return
				}
			}
		}()
	}

	_, err := writer.Write([]byte(`{"success": true }`))
	if err != nil {
		fmt.Println(err.Error())
	}

}
