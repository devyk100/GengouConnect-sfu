package connection

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const PreviousChatAmount = 20

type Chat struct {
	EventType string `json:"eventType"`
	Text      string `json:"text"`
	From      string `json:"from"`
}

type ChatJSON struct {
	Text string `json:"text"`
	From string `json:"from"`
}

type ChatsJSON struct {
	Chats []ChatJSON `json:"chats"`
}

func (e *Class) ChatHandler() {
	for event := range Classes[e.ClassId].Chats {
		if event.EventType == ChatEventType {
			fmt.Println("WRITING EVENTS to STUDENTs CHAT")

			// Tracking the previous chats
			if len(e.PreviousChats) > PreviousChatAmount {
				for len(e.PreviousChats) >= PreviousChatAmount {
					e.PreviousChats = e.PreviousChats[1:]
				}
			}

			e.PreviousChats = append(e.PreviousChats, event)

			if e.Instructor != nil {
				err := e.Instructor.conn.WriteJSON(event)
				if err != nil {
					fmt.Println("ERROR WRITING EVENTS to INSTRUCTOR CHAT")
				}
			}

			for index, learner := range e.Learners {
				fmt.Println("FOUND LEARNERS ", index, event.EventType)
				err := learner.conn.WriteJSON(event)
				if err != nil {
					fmt.Println("Error writing event")
					return
				}
			}

		}

	}
}

func PreviousChatHandler(w http.ResponseWriter, r *http.Request) {
	classId := r.URL.Query().Get("classId")
	userId := r.URL.Query().Get("userId")

	w.Header().Set("Content-Type", "application/json")

	if classId == "" || userId == "" || Classes[classId] != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"success":false}`))
		if err != nil {
			fmt.Println("Error in sending previous chats\n", err.Error())
		}
	}

	var chats []ChatJSON
	for _, chat := range Classes[classId].PreviousChats {
		chats = append(chats, ChatJSON{
			Text: chat.Text,
			From: chat.From,
		})
	}

	w.WriteHeader(http.StatusOK)
	jsonChats, err := json.Marshal(chats)
	if err != nil {
		fmt.Println("Error in sending previous chats\n", err.Error())
	}
	_, err = w.Write(jsonChats)
	if err != nil {
		fmt.Println("Error in sending previous chats\n", err.Error())
		return
	}
	return
}
