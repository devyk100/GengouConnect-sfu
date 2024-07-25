package connection_structs

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
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

type Text struct {
	Text   string `json:"text"`
	Size   int    `json:"size"`
	IsBold bool   `json:"isBold"`
}

type Annotation struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Color     string `json:"color"`
	Thickness int    `json:"thickness"`
}

type Shape struct {
	X1        int    `json:"x1"`
	Y1        int    `json:"y1"`
	X2        int    `json:"x2"`
	Y2        int    `json:"y2"`
	ShapeType string `json:"shapeType"`
	Color     string `json:"color"`
	Thickness int    `json:"thickness"`
}

type Image struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

type WebrtcEvent struct {
	Sdp string `json:"sdp"`
}

type Event struct {
	EventType   string      `json:"eventType"`
	Text        Text        `json:"text"`
	Annotation  Annotation  `json:"annotation"`
	Shape       Shape       `json:"shape"`
	Image       Image       `json:"image"`
	Chat        Chat        `json:"chat"`
	WebrtcEvent WebrtcEvent `json:"webrtcEvent"`
}

const (
	ChatEventType         = "chat"
	PreviousChatEventType = "previousChat"
	BoardEventType
	WebRTCEventType = "webrtc"
)

const PreviousChatAmount = 20

type Chat struct {
	EventType   string      `json:"eventType"`
	WebrtcEvent WebrtcEvent `json:"webrtcEvent"` // TEMPORARILY
	Text        string      `json:"text"`
	From        string      `json:"from"`
}

type ChatJSON struct {
	Text string `json:"text"`
	From string `json:"from"`
}

type ChatsJSON struct {
	Chats []ChatJSON `json:"chats"`
}

type Class struct {
	Instructor    *InstructorClient
	Learners      []*LearnerClient
	IsLive        bool
	ClassId       string
	Events        chan Event
	Chats         chan Chat
	LearnersLock  sync.RWMutex
	ChatsLock     sync.RWMutex
	PreviousChats []Chat
}
type InstructorClient struct {
	Id      string
	ClassId string
	Conn    *websocket.Conn
}

type LearnerClient struct {
	Id      string
	ClassId string
	Conn    *websocket.Conn
}

func (e *Class) ChatHandler() {
	for event := range Classes[e.ClassId].Chats {
		if event.EventType == ChatEventType {
			fmt.Println("WRITING EVENTS to STUDENTs CHAT")

			// Tracking the previous chats
			e.ChatsLock.Lock()
			if len(e.PreviousChats) > PreviousChatAmount {
				for len(e.PreviousChats) >= PreviousChatAmount {
					e.PreviousChats = e.PreviousChats[1:]
				}
			}

			e.PreviousChats = append(e.PreviousChats, event)
			e.ChatsLock.Unlock()

			if e.Instructor != nil {
				err := e.Instructor.Conn.WriteJSON(event)
				if err != nil {
					fmt.Println("ERROR WRITING EVENTS to INSTRUCTOR CHAT")
				}
			}

			for index, learner := range e.Learners {
				fmt.Println("FOUND LEARNERS ", index, event.EventType)
				err := learner.Conn.WriteJSON(event)
				if err != nil {
					fmt.Println("Error writing event")
					return
				}
			}

		}

	}
}

var Classes = map[string]*Class{}

func (e *Class) BoardEventHandler() {
	defer e.LearnersLock.Unlock()
	for event := range Classes[e.ClassId].Events {
		if event.EventType == ChatEventType || event.EventType == PreviousChatEventType {
			return
		}
		fmt.Println("WRITING EVENTS to STUDENTs")
		e.LearnersLock.Lock()
		for index, learner := range e.Learners {
			fmt.Println("FOUND LEARNERS ", index, event.EventType)
			err := learner.Conn.WriteJSON(event)
			if err != nil {
				fmt.Println("Error writing event")
				return
			}
		}
	}
}
