package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/pion/webrtc/v3"
)

func ReadIntoStruct(r io.Reader, v any) error {
	bodyBytes, err := io.ReadAll(r)

	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, v)
}

func writeJsonResponse(w http.ResponseWriter, value any) {
	response, err := json.Marshal(value)

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

type SessionStartReq struct {
	SessionDescription string `json:"sessionDescription"`
}

func startNewPeerConnection(sessionDescription string, ready chan<- string) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	closeGoRoutine := make(chan bool)

	peerConnection, err := webrtc.NewPeerConnection(config)
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
			closeGoRoutine <- true
		}
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {

		label := d.Label()

		// Register channel opening handling
		d.OnOpen(func() {
			fmt.Print("Data channel has been opened: ", label)
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel: '%s': '%s'\n", label, string(msg.Data))
		})

		d.OnClose(func() {
			fmt.Printf("Closing data channel: %s", label)
		})
	})

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

	select {
	case <-closeGoRoutine:
		log.Print("Closing goroutine")
		return
	}
}

func HandleSessionStart(w http.ResponseWriter, req *http.Request) {
	var sessionData SessionStartReq
	err := ReadIntoStruct(req.Body, &sessionData)

	if err != nil {
		panic(err)
	}

	ready := make(chan string)
	go startNewPeerConnection(sessionData.SessionDescription, ready)

	localDescription := <-ready
	writeJsonResponse(w, SessionStartReq{localDescription})

	w.WriteHeader(http.StatusOK)

}

func HandleNewGame(w http.ResponseWriter, req *http.Request) {
	game := CreateGame()

	type NewGameResponse struct {
		Id string `json:"gameId"`
	}

	writeJsonResponse(w, NewGameResponse{game.Id})
	w.WriteHeader(http.StatusCreated)
}

func HandleStatic(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join("public", path))
	})
}

func HandleValidatedStatic(
	path string,
	validator func(w http.ResponseWriter, req *http.Request) bool,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !validator(w, req) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			http.ServeFile(w, req, filepath.Join("public", path))

		}
	})
}

func ValidateGameId(w http.ResponseWriter, req *http.Request) bool {
	gameId := chi.URLParam(req, "gameId")
	log.Println("gameId...", gameId)
	_, ok := GetGame(gameId)
	return ok
}
