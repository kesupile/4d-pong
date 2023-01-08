package games

import (
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type Player struct {
	PeerConnection *webrtc.PeerConnection
	DataChannel    *webrtc.DataChannel
	Position       string
	Coordinates    *[2]int
	Dimensions     *[2]int
	IsActive       bool
	MagX           int
	MagY           int
}

type GameEvent struct {
	Type string
	Data any
}

type Game struct {
	Id                  string
	NPlayers            int
	NPlayersConnected   int
	StatusUpdatesActive bool
	Active              bool
	Width               int
	Height              int
	TopPlayer           *Player
	BottomPlayer        *Player
	LeftPlayer          *Player
	RightPlayer         *Player
	events              chan GameEvent
	stopStatusUpdates   chan bool
}

func (game *Game) IsAcceptingConnections() bool {
	return game.NPlayersConnected < game.NPlayers
}

func (game *Game) FindPlayerToAssign() (*Player, error) {
	if !game.IsAcceptingConnections() {
		return nil, errors.New("maximum players reached")
	}

	player := &Player{}

	// TODO: Left and right players
	baseWidth := 50
	baseHeight := 10
	middleXPosition := (game.Width / 2) - baseWidth/2

	var topLeftCoordinates [2]int
	var dimensions [2]int

	switch {
	case game.TopPlayer == nil:
		player.Position = "top"

		topLeftCoordinates[0] = middleXPosition
		topLeftCoordinates[1] = 0
		player.Coordinates = &topLeftCoordinates

		dimensions[0] = baseWidth
		dimensions[1] = baseHeight
		player.Dimensions = &dimensions

		player.MagX = baseWidth
		player.MagY = baseHeight

		game.TopPlayer = player
	case game.BottomPlayer == nil:
		player.Position = "bottom"
		topLeftCoordinates[0] = middleXPosition
		topLeftCoordinates[1] = game.Height - baseHeight
		player.Coordinates = &topLeftCoordinates

		dimensions[0] = baseWidth
		dimensions[1] = baseHeight
		player.Dimensions = &dimensions

		player.MagX = baseWidth
		player.MagY = baseHeight

		game.BottomPlayer = player
	case game.LeftPlayer == nil:
		player.Position = "left"
		game.LeftPlayer = player
	default:
		player.Position = "right"
		game.RightPlayer = player
	}

	return player, nil
}

var gameStore = map[string]*Game{}

func CreateGame() *Game {
	game := &Game{
		Id:                uuid.NewString(),
		Active:            false,
		Width:             200,
		Height:            200,
		NPlayers:          2,
		NPlayersConnected: 0,
		events:            make(chan GameEvent),
		stopStatusUpdates: make(chan bool),
	}

	go game.listenForEvents()
	gameStore[game.Id] = game
	return game
}

func GetGame(id string) (*Game, bool) {
	game, ok := gameStore[id]
	return game, ok
}

func handleDataChannelOpen(game *Game, player *Player) func() {
	return func() {
		game.events <- GameEvent{Type: ACTIVATE_PLAYER, Data: player.Position}
		game.events <- GameEvent{Type: START_STATUS_UPDATES}
	}
}

func (game *Game) listenForEvents() {
listener:
	for {
		event := <-game.events

		yellowColour := "\033[33m"
		resetColour := "\033[0m"

		log.Println(string(yellowColour), "New event", event.Type, string(resetColour))

		switch event.Type {
		case CLEAN_UP_CONNECTION:
			cleanUpConnection(game, event.Data.(string))
		case ACTIVATE_PLAYER:
			activatePlayer(game, event.Data.(string))
		case START_STATUS_UPDATES:
			go startStatusUpdates(game)
		case STOP_LISTENING:
			break listener
		}
	}
}

func attachDataChannelHandlers(game *Game, player *Player) {
	dataChannel := player.DataChannel
	label := dataChannel.Label()

	// Register channel opening handling
	dataChannel.OnOpen(handleDataChannelOpen(game, player))

	// Register text message handling
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Message from DataChannel: '%s': '%s'\n", label, string(msg.Data))
	})

	dataChannel.OnClose(func() {
		fmt.Printf("Closing data channel: %s", label)
		game.events <- GameEvent{Type: CLEAN_UP_CONNECTION, Data: player.Position}
	})
}

func registerDataChannel(game *Game, player *Player, dataChannel *webrtc.DataChannel) error {
	player.DataChannel = dataChannel
	attachDataChannelHandlers(game, player)
	return nil
}

func attachPeerConnectionHandlers(game *Game, player *Player, peerConnection *webrtc.PeerConnection) {
	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateDisconnected {
			game.events <- GameEvent{Type: CLEAN_UP_CONNECTION, Data: player.Position}
		}

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
		}
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		registerDataChannel(game, player, dataChannel)
	})

}

func RegisterPeerConnection(gameId string, peerConnection *webrtc.PeerConnection) error {
	game, ok := GetGame(gameId)
	if !ok {
		return errors.New("no such game")
	}

	player, err := game.FindPlayerToAssign()

	if err != nil {
		return err
	}

	player.PeerConnection = peerConnection
	attachPeerConnectionHandlers(game, player, peerConnection)
	return nil
}
