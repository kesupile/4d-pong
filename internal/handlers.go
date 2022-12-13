package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pion/webrtc/v3"
)

func ReadIntoStruct(r io.Reader, v any) error {
	bodyBytes, err := io.ReadAll(r)

	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, v)
}

type SessionStartReq struct {
	SessionDescription string `json:"sessionDescription"`
}

func startNewPeerConnection(sessionDescription string, ready chan<- string) {
	s := webrtc.SettingEngine{}
	s.DetachDataChannels()

	api := webrtc.NewAPI((webrtc.WithSettingEngine(s)))

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("Cannot close peerConnection %v\n", cErr)
		}
	}()

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}
	})

	// TODO: handle data channel data

	sessionDescriptionBytes, sessionDecodingErr := base64.StdEncoding.DecodeString(sessionDescription)
	if sessionDecodingErr != nil {
		panic(sessionDecodingErr)
	}

	offer := webrtc.SessionDescription{}
	offerUnmarshalError := json.Unmarshal(sessionDescriptionBytes, &offer)
	if offerUnmarshalError != nil {
		panic(offerUnmarshalError)
	}

	// Set the remote SessionDescription
	setRemoteDescriptionErr := peerConnection.SetRemoteDescription(offer)
	if setRemoteDescriptionErr != nil {
		panic(setRemoteDescriptionErr)
	}

	// Create answer
	answer, createAnswerErr := peerConnection.CreateAnswer(nil)
	if createAnswerErr != nil {
		panic(createAnswerErr)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription and starts ou UDP listeners
	setLocalDescriptionErr := peerConnection.SetLocalDescription(answer)
	if setLocalDescriptionErr != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disable trickle ICE
	// we do this because we only can exchange on signalling message
	// in a production application you should exchange ICECandidates via OnICECandidate
	// TODO: WTF does this mean? lol
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	outputAnswerBytes, outputAnswerErr := json.Marshal(*peerConnection.LocalDescription())
	if outputAnswerErr != nil {
		panic(outputAnswerErr)
	}

	localDescription := base64.StdEncoding.EncodeToString(outputAnswerBytes)

	ready <- localDescription

	select {}
}

type HandlerStruct struct{}

var HandleSessionStart = HandlerStruct{}

func (handler HandlerStruct) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	var sessionData SessionStartReq
	err := ReadIntoStruct(request.Body, &sessionData)

	if err != nil {
		panic(err)
	}

	ready := make(chan string)
	go startNewPeerConnection(sessionData.SessionDescription, ready)

	localDescription := <-ready
	fmt.Println("localDescription", localDescription)

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)

	response, err := json.Marshal(SessionStartReq{localDescription})

	if err != nil {
		panic(err)
	}

	responseWriter.Write(response)
}
