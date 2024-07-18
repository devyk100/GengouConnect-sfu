package webrtc_sfu

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc/v4"
)

func encodeSDP(answer *webrtc.SessionDescription) string {
	jsonSdpString, err := json.Marshal(answer)
	if err != nil {
		fmt.Println(err.Error(), "At encoding the sdp to string json")
	}
	return base64.StdEncoding.EncodeToString(jsonSdpString)
}
