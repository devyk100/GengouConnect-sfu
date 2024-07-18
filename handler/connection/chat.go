package connection

import (
	"encoding/json"
	"fmt"
	"net/http"
	connection_structs "sfu/handler/connection-structs"
)

func PreviousChatHandler(w http.ResponseWriter, r *http.Request) {
	classId := r.URL.Query().Get("classId")
	userId := r.URL.Query().Get("userId")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin domain
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if classId == "" || userId == "" || connection_structs.Classes[classId] == nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"success":false}`))
		if err != nil {
			fmt.Println("Error in sending previous chats\n", err.Error())
		}
		return
	}

	var chats []connection_structs.ChatJSON
	for _, chat := range connection_structs.Classes[classId].PreviousChats {
		chats = append(chats, connection_structs.ChatJSON{
			Text: chat.Text,
			From: chat.From,
		})
	}

	jsonChats, err := json.Marshal(chats)
	if err != nil {
		fmt.Println("Error in sending previous chats\n", err.Error())
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonChats)
	if err != nil {
		fmt.Println("Error in sending previous chats\n", err.Error())
		return
	}
	return
}
