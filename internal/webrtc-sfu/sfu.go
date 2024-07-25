package webrtc_sfu

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	pionInterceptor "github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
	"log"
	"os"
	connection_structs "sfu/handler/connection-structs"
)

type LiveClass struct {
	InstructorPeerConnection *webrtc.PeerConnection
	LocalTrack               *webrtc.TrackLocalStaticRTP
	LocalAudioTrack          *webrtc.TrackLocalStaticRTP
	LearnerPeerConnections   []*webrtc.PeerConnection
	ClassId                  string
	// Add RW mutexes here
}

var LiveClasses map[string]*LiveClass = make(map[string]*LiveClass)

func InitializeSfuConnection(event connection_structs.Event, client *websocket.Conn, userType string, classId string) *webrtc.PeerConnection {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	turnIp := os.Getenv("PUBLIC_IP")
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs:           []string{"stun:" + turnIp + ":3478", "stun:stun.l.google.com:19302", "turn:" + turnIp + ":3478"},
				Username:       "user",
				Credential:     "pass",
				CredentialType: 0,
			},
		},
	}
	if userType == connection_structs.Instructor {
		offer := webrtc.SessionDescription{}
		decodeSDP(&event, &offer)
		fmt.Println("Decoded the SDP and proceeding with the instructor broadcast")

		mediaEngine := webrtc.MediaEngine{}
		err := mediaEngine.RegisterDefaultCodecs()
		if err != nil {
			fmt.Println(err.Error(), "In registering the default codecs")
		}

		interceptorRegistry := &pionInterceptor.Registry{}

		err = webrtc.RegisterDefaultInterceptors(&mediaEngine, interceptorRegistry)
		if err != nil {
			fmt.Println(err.Error(), "In registering the default interceptors")
		}

		intervalPliFactory, err := intervalpli.NewReceiverInterceptor()

		if err != nil {
			fmt.Println(err.Error(), "In in interval pli interceptorRegistry")
		}

		interceptorRegistry.Add(intervalPliFactory)

		peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine), webrtc.WithInterceptorRegistry(interceptorRegistry)).NewPeerConnection(peerConnectionConfig)

		if err != nil {
			fmt.Println(err.Error(), "In creating new peer connection obj")
		}

		_, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
		if err != nil {
			fmt.Println(err.Error(), "In adding video transceiver")
		}

		if LiveClasses[classId] == nil {
			LiveClasses[classId] = &LiveClass{
				InstructorPeerConnection: peerConnection,
				LocalTrack:               nil,
				LearnerPeerConnections:   []*webrtc.PeerConnection{},
				ClassId:                  classId,
			}
		} else {
			LiveClasses[classId].InstructorPeerConnection = peerConnection
			LiveClasses[classId].LocalTrack = nil
			LiveClasses[classId].ClassId = classId
		}
		LiveClasses[classId].HandleBroadCast()

		err = peerConnection.SetRemoteDescription(offer)
		if err != nil {
			fmt.Println(err.Error(), "In setting remote description")
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			fmt.Println(err.Error(), "In creating answer")
		}

		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

		err = peerConnection.SetLocalDescription(answer)

		if err != nil {
			fmt.Println(err.Error(), "In setting local description")
		}

		<-gatherComplete

		sdpString := encodeSDP(peerConnection.LocalDescription())

		err = client.WriteJSON(&connection_structs.Event{
			EventType: connection_structs.WebRTCEventType,
			WebrtcEvent: connection_structs.WebrtcEvent{
				Sdp: sdpString,
			},
		})
		if err != nil {
			fmt.Println(err.Error(), "In writing SDP")
		}

		//go LiveClasses[classId].HandleBroadcastLearners()
		return peerConnection
	} else {
		receiveOnlyOffer := webrtc.SessionDescription{}
		decodeSDP(&event, &receiveOnlyOffer)
		peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
		if err != nil {
			fmt.Println(err.Error(), "In creating new peer connection obj")
		}
		if LiveClasses[classId] == nil {
			fmt.Println("No class was found yet")
			return nil
		} else if LiveClasses[classId].LocalTrack == nil {
			fmt.Println("No local track found yet from the instructor")
			return nil
		}
		rtpSender, err := peerConnection.AddTrack(LiveClasses[classId].LocalTrack)

		if err != nil {
			fmt.Println(err.Error(), "In adding track")
		}

		_, err = peerConnection.AddTrack(LiveClasses[classId].LocalAudioTrack)
		if err != nil {
			fmt.Println(err.Error(), "In adding audio track")
		}
		go func() {
			rtcpBuf := make([]byte, 1000)
			for {
				_, _, rtcpErr := rtpSender.Read(rtcpBuf)
				if LiveClasses[classId] == nil {
					fmt.Println("Classes are closed")
					return
				}
				if rtcpErr != nil {
					fmt.Println(rtcpErr.Error(), "In reading packets from the receivers")
					return
				}
			}
		}()

		err = peerConnection.SetRemoteDescription(receiveOnlyOffer)
		if err != nil {
			fmt.Println(err.Error(), "In setting remote description")
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			fmt.Println(err.Error(), "In creating answer")
		}

		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			fmt.Println(err.Error(), "In setting local description")
		}

		<-gatherComplete

		sdpString := encodeSDP(peerConnection.LocalDescription())

		err = client.WriteJSON(connection_structs.Event{
			EventType: connection_structs.WebRTCEventType,
			WebrtcEvent: connection_structs.WebrtcEvent{
				Sdp: sdpString,
			},
		})
		if err != nil {
			fmt.Println(err.Error(), "In writing SDP")
		}
		return peerConnection
	}

}
