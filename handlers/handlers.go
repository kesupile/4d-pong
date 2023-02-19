package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	games "server/game"

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
	PlayerName         string `json:"playerName"`
}

func startNewPeerConnection(sessionStartReq SessionStartReq, gameId string, ready chan<- string) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	err = games.RegisterPeerConnection(gameId, peerConnection, sessionStartReq.PlayerName)

	if err != nil {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("Cannot close peerConnection %v\n", cErr)
		}
	}

	sessionDescriptionBytes, sessionDecodingErr := base64.StdEncoding.DecodeString(sessionStartReq.SessionDescription)
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
}

func getGameFromRequest(req *http.Request) *games.Game {
	gameId := chi.URLParam(req, "gameId")
	game, _ := games.GetGame(gameId)
	return game
}

func HandleGameStatusGET(w http.ResponseWriter, req *http.Request) {
	game := getGameFromRequest(req)

	type Response struct {
		Active               bool `json:"active"`
		AcceptingConnections bool `json:"acceptingConnections"`
		Height               int  `json:"height"`
		Width                int  `json:"width"`
	}

	writeJsonResponse(w, Response{
		Active:               game.IsActive,
		AcceptingConnections: game.IsAcceptingConnections(),
		Height:               game.Height,
		Width:                game.Width,
	})
	w.WriteHeader(http.StatusOK)
}

func HandleGameJoinPOST(w http.ResponseWriter, req *http.Request) {
	game := getGameFromRequest(req)

	if !game.IsAcceptingConnections() {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var sessionData SessionStartReq
	err := ReadIntoStruct(req.Body, &sessionData)

	if err != nil {
		panic(err)
	}

	ready := make(chan string)
	go startNewPeerConnection(sessionData, game.Id, ready)

	localDescription := <-ready
	type SessionStartResp struct {
		SessionDescription string `json:"sessionDescription"`
	}
	writeJsonResponse(w, SessionStartResp{localDescription})

	w.WriteHeader(http.StatusOK)

}

func HandleNewGamePOST(w http.ResponseWriter, req *http.Request) {
	var gameInfo games.NewGameInfo
	err := ReadIntoStruct(req.Body, &gameInfo)

	if err != nil {
		panic(err)
	}

	game := games.CreateGame(gameInfo)

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

func HandleValidatedRestEndpoint(
	validator func(w http.ResponseWriter, req *http.Request) bool,
	handler func(w http.ResponseWriter, req *http.Request),
) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ok := validator(w, req)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			handler(w, req)
		}
	})
}

func ValidateGameId(w http.ResponseWriter, req *http.Request) bool {
	gameId := chi.URLParam(req, "gameId")
	_, ok := games.GetGame(gameId)
	return ok
}
