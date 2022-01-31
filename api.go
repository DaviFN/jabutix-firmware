package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"

	"github.com/pion/mediadevices"
	//"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
	//"github.com/pion/mediadevices/pkg/driver/camera"
	//"github.com/pion/mediadevices/pkg/codec/opus" // This is required to use opus audio encoder
	"github.com/pion/mediadevices/pkg/codec/x264" // This is required to use h264 video encoder
	//Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	//"github.com/pion/mediadevices/pkg/driver/videotest"
	//"github.com/pion/mediadevices/pkg/driver/audiotest"

	_ "github.com/pion/mediadevices/pkg/driver/camera" // This is required to register camera adapter

	"encoding/json"
	"math/rand"

	//"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var config webrtc.Configuration = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

var msgtype_sdpOffer byte = 'a'
var msgtype_sdpAnswer byte = 'b'

var msgtype_commandMoveForward byte = 'c'
var msgtype_commandMoveLeft byte = 'd'
var msgtype_commandMoveRight byte = 'e'

var msgtype_commandGetDiscreteMovementConfig byte = 'f'
var msgtype_commandSetDiscreteMovementConfig byte = 'g'

var msgtype_callbackMovedForward byte = 'h'
var msgtype_callbackMovedLeft byte = 'i'
var msgtype_callbackMovedRight byte = 'j'


func ApiHandler(w http.ResponseWriter, r *http.Request) {

	var handlerId int = rand.Intn(10000)
	fmt.Printf("websocketHandler called, ID = %d\n", handlerId)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("upgrader.Upgrade error: %s\n", err)
		return
	}
	fmt.Printf("%s connected to server via WebSockets\n", conn.RemoteAddr())

	for {
		
		fmt.Printf("about to read message from browser...\n")
		
		// Read message from browser
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("conn.ReadMessage error: %s\n", err)
			return
		}

		// Print the message to the console
		fmt.Printf("(%d) message from %s\n", msgType, conn.RemoteAddr())
		//fmt.Printf("(%d) message from %s: %s\n", msgType, conn.RemoteAddr(), string(msg))

		// serving the message accordingly
		var msgtype byte = msg[0]
		fmt.Printf("msgtype: %c\n", msgtype)
		
		
		
		// API handlers for getting and setting movement configs
		
		
		if msgtype == msgtype_commandGetDiscreteMovementConfig {
			var discreteMovementConfigMessage string = GetDiscreteMovementConfigMessageString()
			
			fmt.Printf("sending discreteMovementConfig message\n")
			fmt.Printf("which is: %s\n", discreteMovementConfigMessage)
			
			if err = conn.WriteMessage(msgType, []byte(discreteMovementConfigMessage)); err != nil {
				fmt.Printf("conn.WriteMessage error: %s\n", err)
				return
			}
			
		}
		
		if msgtype == msgtype_commandSetDiscreteMovementConfig {
			fmt.Printf("command for setting movement configs received\nmessage: %s", msg)
			
			SetDiscreteMovementConfigBasedOnMessage(string(msg));
			
		}
		
		
		
		// API handlers for commands related to moving the jabuti
		
		if msgtype == msgtype_commandMoveForward {
			MoveForward()
			
			var callbackMsg []byte;
			callbackMsg = append(callbackMsg, msgtype_callbackMovedForward)
			conn.WriteMessage(msgType, []byte(callbackMsg))
			fmt.Printf("sent msgtype_callbackMovedForward\n", err)
			
	
		}
		
		if msgtype == msgtype_commandMoveLeft {
			MoveLeft()
			
			var callbackMsg []byte;
			callbackMsg = append(callbackMsg, msgtype_callbackMovedLeft)
			conn.WriteMessage(msgType, []byte(callbackMsg))
			fmt.Printf("sent msgtype_callbackMovedLeft\n", err)
			
		}
		
		if msgtype == msgtype_commandMoveRight {
			MoveRight()
			
			var callbackMsg []byte;
			callbackMsg = append(callbackMsg, msgtype_callbackMovedRight)
			conn.WriteMessage(msgType, []byte(callbackMsg))
			fmt.Printf("sent msgtype_callbackMovedRight\n", err)
			
		}
		
		
		
		
		// API handlers for stablishing a WebRTC connection (required for the camera)

		if msgtype == msgtype_sdpOffer {
			fmt.Printf("msgtype_sdpOffer detected")

			var offerJSON []byte = msg[1:]

			var offer webrtc.SessionDescription

			err = json.Unmarshal(offerJSON, &offer)
			if err != nil {
				panic(err)
			}
			
			x264Params, err := x264.NewParams()
			if err != nil {
				panic(err)
			}
			x264Params.BitRate = 500_000 // 500kbps
			
			//fmt.Printf("deb 1\n")

			
			//fmt.Printf("deb 2\n")

			codecSelector := mediadevices.NewCodecSelector(
				mediadevices.WithVideoEncoders(&x264Params),
			)
			
			//fmt.Printf("deb 3\n")

			mediaEngine := webrtc.MediaEngine{}
			
			//fmt.Printf("deb 4\n")
			
			codecSelector.Populate(&mediaEngine)
			
			//fmt.Printf("deb 5\n")

			api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
			
			//fmt.Printf("deb 6\n")

			peerConnection, err := api.NewPeerConnection(config)
			if err != nil {
				panic(err)
			}
			
			//fmt.Printf("deb 7\n")

			// Set the handler for ICE connection state
			// This will notify you when the peer has connected/disconnected
			peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
				fmt.Printf("Connection State has changed %s \n", connectionState.String())
			})
			
			//fmt.Printf("deb 8\n")
			
			s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
				Video: func(c *mediadevices.MediaTrackConstraints) {
					//c.FrameFormat = prop.FrameFormat(frame.FormatI420)
					c.Width = prop.Int(640)
					c.Height = prop.Int(480)
				},
				Codec: codecSelector,
			})
			if err != nil {
				panic(err)
			}
			
			//fmt.Printf("deb 9\n")
			
			track := s.GetVideoTracks()[0]
			videoTrack := track.(*mediadevices.VideoTrack)
			//defer videoTrack.Close()
			
			_, err = peerConnection.AddTransceiverFromTrack(videoTrack,
				webrtc.RtpTransceiverInit{
					Direction: webrtc.RTPTransceiverDirectionSendonly,
				},
				)
				if err != nil {
					panic(err)
				}
			
			for _, track := range s.GetTracks() {
				track.OnEnded(func(err error) {
				fmt.Printf("Track (ID: %s) ended with error: %v\n",
					track.ID(), err)
				})
				
			//fmt.Printf("deb 10\n")

				//_, err = peerConnection.AddTransceiverFromTrack(track,
				//webrtc.RtpTransceiverInit{
				//	Direction: webrtc.RTPTransceiverDirectionSendonly,
				//},
				//)
				//if err != nil {
				//	panic(err)
				//}
				
				//fmt.Printf("deb 11\n")
				
			}

			// Set the remote SessionDescription
			err = peerConnection.SetRemoteDescription(offer)
			if err != nil {
				panic(err)
			}
			
			//fmt.Printf("deb 12\n")

			// Create an answer
			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				panic(err)
			}
			
			//fmt.Printf("deb 13\n")

			// Create channel that is blocked until ICE Gathering is complete
			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
			
			//fmt.Printf("deb 14\n")
			

			
			
			//fmt.Printf("deb 15\n")
			
			
			
			//fmt.Printf("deb 16\n")
			
			//fmt.Printf("+v", answer)
			
			// Sets the LocalDescription, and starts our UDP listeners
			err = peerConnection.SetLocalDescription(answer)
			if err != nil {
				fmt.Printf("err from SetLocalDescription is not nil\n")
				panic(err)
			}
			
			//fmt.Printf("sleeping\n")
			//time.Sleep(5 * time.Second)
			//fmt.Printf("slept\n")
			
			//fmt.Printf("deb 17\n")

			// Block until ICE Gathering is complete, disabling trickle ICE
			// we do this because we only can exchange one signaling message
			// in a production application you should exchange ICE Candidates via OnICECandidate
			<-gatherComplete
			
			//fmt.Printf("deb 18\n")

			fmt.Printf("gather complete\n")

			answerJSON, err := json.Marshal(peerConnection.LocalDescription())

			var answerMessage string = string(msgtype_sdpAnswer) + string(answerJSON)

			//fmt.Printf("answerMessage: %s\n", answerMessage)
			fmt.Printf("generated answer message\n")
			
			fmt.Printf("sending answer message\n")
			if err = conn.WriteMessage(msgType, []byte(answerMessage)); err != nil {
				fmt.Printf("conn.WriteMessage error: %s\n", err)
				return
			}
			

			
			// Block forever
			//select {}

		}

		// Write message back to browser
		//if err = conn.WriteMessage(msgType, msg); err != nil {
		//	fmt.Printf("conn.WriteMessage error: %s\n", err)
		//	return
		//}

	}

}