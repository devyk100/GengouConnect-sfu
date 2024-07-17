package connection

//yash@Om#12$
//yash@135
//ommehta@123

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Class struct {
	Instructor    *InstructorClient
	Learners      []*LearnerClient
	isLive        bool
	ClassId       string
	Events        chan BoardEvent
	Chats         chan Chat
	LearnersLock  sync.RWMutex
	ChatsLock     sync.RWMutex
	PreviousChats []Chat
}

var Classes = map[string]*Class{}

func CreateClassHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("CreateClassHandler Called")
	classId := request.URL.Query().Get("classId")

	// Make a check if this was put in the Database or not by the backend

	writer.Header().Set("Content-Type", "application/json")
	if classId == "" || Classes[classId] != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, err := writer.Write([]byte(`{"success": false }`))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		return
	}
	if Classes[classId] == nil {
		Classes[classId] = &Class{
			Instructor:    nil,
			Learners:      nil,
			isLive:        true,
			Events:        make(chan BoardEvent),
			Chats:         make(chan Chat, 50),
			ClassId:       classId,
			ChatsLock:     sync.RWMutex{},
			PreviousChats: make([]Chat, PreviousChatAmount),
			LearnersLock:  sync.RWMutex{},
		}
	}

	writer.WriteHeader(http.StatusCreated)

	go func(classId string) {
		for {
			select {
			case <-time.After(time.Minute * 60):
				if Classes[classId].Instructor == nil {
					fmt.Println("Delete the class room", classId)
					close(Classes[classId].Events)
					close(Classes[classId].Chats)
					Classes[classId] = nil
					return
				}
			}
		}
	}(classId)

	go Classes[classId].BoardEventHandler()
	go Classes[classId].ChatHandler()

	_, err := writer.Write([]byte(`{"success": true }`))
	if err != nil {
		fmt.Println(err.Error())
	}

}
