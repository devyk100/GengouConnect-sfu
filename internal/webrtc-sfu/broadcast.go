package webrtc_sfu

import (
	"fmt"
	"github.com/pion/webrtc/v4"
)

func (instance *LiveClass) HandleBroadCast() {
	instance.InstructorPeerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		trackType := remoteTrack.Kind().String()
		if trackType == "video" {
			localTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", instance.ClassId)
			if newTrackErr != nil {
				fmt.Println("Error creating new track", newTrackErr.Error())
			}

			instance.LocalTrack = localTrack

			rtpBuffer := make([]byte, 2000)
			for {
				i, _, readErr := remoteTrack.Read(rtpBuffer)
				if readErr != nil {
					fmt.Println("Error reading from remote video track", readErr.Error())
					return
				}

				_, err := localTrack.Write(rtpBuffer[:i])
				if err != nil {
					fmt.Println("Error writing to local video track", err.Error())
					return
				}
			}
		} else if trackType == "audio" {
			fmt.Println("AUDIO TRACK ADDED YES!")
			localAudioTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "audio", instance.ClassId)
			if newTrackErr != nil {
				fmt.Println("Error creating new track", newTrackErr.Error())
			}

			instance.LocalAudioTrack = localAudioTrack

			rtpBuffer := make([]byte, 4000)

			for {
				i, _, readErr := remoteTrack.Read(rtpBuffer)
				if readErr != nil {
					fmt.Println("Error reading from remote audio track", readErr.Error())
					return
				}

				_, err := localAudioTrack.Write(rtpBuffer[:i])
				if err != nil {
					fmt.Println("Error writing to local audio track", err.Error())
					return
				}
			}
		}

	})
}
