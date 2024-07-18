package webrtc_sfu

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc/v4"
	connection_structs "sfu/handler/connection-structs"
)

func decodeSDP(event *connection_structs.Event, obj *webrtc.SessionDescription) {
	jsonSdpString, err := base64.StdEncoding.DecodeString(event.WebrtcEvent.Sdp)
	if err != nil {
		fmt.Println("decode sdp error:", err)
	}
	if err = json.Unmarshal(jsonSdpString, obj); err != nil {
		fmt.Println("sdp unmarshal to struct error:", err)
	}
}
