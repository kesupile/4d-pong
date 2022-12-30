package games

import (
	"encoding/json"
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
}

type GameEvent struct {
	Type string
	Data any
}

type Game struct {
	Id                string
	NPlayers          int
	NPlayersConnected int
	Active            bool
	Width             int
	Height            int
	TopPlayer         *Player
	BottomPlayer      *Player
	LeftPlayer        *Player
	RightPlayer       *Player
	events            chan GameEvent
}

func (game *Game) IsAcceptingConnections() bool {
	fmt.Printf("N players %v, Connected %v \n", game.NPlayers, game.NPlayersConnected)
	return game.NPlayersConnected < game.NPlayers
}

func (game *Game) FindPlayerToAssign() (*Player, error) {
	if !game.IsAcceptingConnections() {
		return nil, errors.New("maximum players reached")
	}

	player := &Player{}

	baseWidth := 60
	baseHeight := 20
	x2 := (game.Width / 2) - baseWidth/2

	// The centre point of the player
	var pos [2]int

	// The width and height of the representing player element
	var dim [2]int

	switch {
	case game.TopPlayer == nil:
		player.Position = "top"

		pos[0] = x2
		pos[1] = 0
		player.Coordinates = &pos

		dim[0] = baseWidth
		dim[1] = baseHeight
		player.Dimensions = &dim

		game.TopPlayer = player
	case game.BottomPlayer == nil:
		player.Position = "bottom"
		pos[0] = x2
		pos[1] = game.Height - baseHeight
		player.Coordinates = &pos

		dim[0] = baseWidth
		dim[1] = baseHeight
		player.Dimensions = &dim

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
	game := Game{
		Id:                uuid.NewString(),
		Active:            false,
		Width:             200,
		Height:            200,
		NPlayers:          2,
		NPlayersConnected: 0,
		events:            make(chan GameEvent),
	}

	go game.listenForEvents()
	gameStore[game.Id] = &game
	return &game
}

func GetGame(id string) (*Game, bool) {
	game, ok := gameStore[id]
	return game, ok
}

func handleDataChannelOpen(game *Game, player *Player) func() {
	return func() {
		dataChannel := player.DataChannel
		fmt.Print("Data channel has been opened: ", dataChannel.Label())
		game.NPlayersConnected += 1

		type InitMessage struct {
			Type              string `json:"type"`
			PlayerPosition    string `json:"playerPosition"`
			PlayerCoordinates [2]int `json:"playerCoordinates"`
			PlayerDimensions  [2]int `json:"playerDimensions"`
			Height            int    `json:"height"`
			Width             int    `json:"width"`
		}

		message, _ := json.Marshal(InitMessage{
			Type:              "init",
			PlayerPosition:    player.Position,
			PlayerCoordinates: *player.Coordinates,
			PlayerDimensions:  *player.Dimensions,
			Height:            game.Height,
			Width:             game.Width,
		})

		dataChannel.SendText(string(message))
	}
}

func (game *Game) listenForEvents() {
listener:
	for {
		event := <-game.events

		log.Println("New event", event.Type)

		switch event.Type {
		case CLEAN_UP_CONNECTION:
			cleanUpConnection(game, event.Data.(string))
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
