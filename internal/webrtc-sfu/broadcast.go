package webrtc_sfu

import (
	"fmt"
	"github.com/pion/webrtc/v4"
)

func (instance *LiveClass) HandleBroadCast() {
	instance.InstructorPeerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		localTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(remoteTrack.Codec().RTPCodecCapability, "video", instance.ClassId)
		if newTrackErr != nil {
			fmt.Println("Error creating new track", newTrackErr.Error())
		}

		instance.LocalTrack = localTrack

		rtpBuffer := make([]byte, 2000)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuffer)
			if readErr != nil {
				fmt.Println("Error reading from remote track", readErr.Error())
			}

			_, err := localTrack.Write(rtpBuffer[:i])
			if err != nil {
				fmt.Println("Error writing to local track", err.Error())
			}
		}
	})
}

func (instance *LiveClass) HandleBroadcastLearners() {

}
