package games

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type Ball struct {
	IsVisible      bool
	CentrePosition *[2]float32
	Velocity       *[2]float32
	Radius         float32
}

type Player struct {
	Name           string
	PeerConnection *webrtc.PeerConnection
	DataChannel    *webrtc.DataChannel
	Position       string
	Coordinates    *[2]int
	IsActive       bool
	MagX           int
	MagY           int
	IsEjected      bool
	Velocity       float32
	ResetVelocity  func()
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
	CreatorName         string
	IsActive            bool
	StartTime           string
	Width               int
	Height              int
	TopPlayer           *Player
	BottomPlayer        *Player
	LeftPlayer          *Player
	RightPlayer         *Player
	events              chan GameEvent
	StopStatusUpdates   chan bool
	StopGame            chan bool
	Balls               *[1]*Ball
}

type NewGameInfo struct {
	NPlayers    int    `json:"nPlayers"`
	CreatorName string `json:"creatorName"`
}

func (game *Game) IsAcceptingConnections() bool {
	return game.StartTime == "" && game.NPlayersConnected < game.NPlayers
}

func (game *Game) FindPlayerToAssign() (*Player, error) {
	if !game.IsAcceptingConnections() {
		return nil, errors.New("maximum players reached")
	}

	player := &Player{}

	topLeftXForTopAndBottomPlayers := (game.Width / 2) - PLAYER_WIDTH/2
	topLeftYForLeftAndRightPlayers := (game.Height / 2) - PLAYER_WIDTH/2

	var topLeftCoordinates [2]int

	switch {
	case game.TopPlayer == nil:
		player.Position = "top"

		player.MagX = PLAYER_WIDTH
		player.MagY = PLAYER_HEIGHT

		topLeftCoordinates[0] = topLeftXForTopAndBottomPlayers
		topLeftCoordinates[1] = 0
		player.Coordinates = &topLeftCoordinates

		game.TopPlayer = player
	case game.BottomPlayer == nil:
		player.Position = "bottom"

		player.MagX = PLAYER_WIDTH
		player.MagY = PLAYER_HEIGHT

		topLeftCoordinates[0] = topLeftXForTopAndBottomPlayers
		topLeftCoordinates[1] = game.Height - player.MagY
		player.Coordinates = &topLeftCoordinates

		game.BottomPlayer = player
	case game.LeftPlayer == nil:
		player.Position = "left"

		player.MagX = PLAYER_HEIGHT
		player.MagY = PLAYER_WIDTH

		topLeftCoordinates[0] = 0
		topLeftCoordinates[1] = topLeftYForLeftAndRightPlayers
		player.Coordinates = &topLeftCoordinates

		game.LeftPlayer = player
	default:
		player.Position = "right"

		player.MagX = PLAYER_HEIGHT
		player.MagY = PLAYER_WIDTH

		topLeftCoordinates[0] = game.Width - player.MagX
		topLeftCoordinates[1] = topLeftYForLeftAndRightPlayers
		player.Coordinates = &topLeftCoordinates

		game.RightPlayer = player
	}

	player.ResetVelocity = WithDebounce(time.Duration((time.Millisecond * 80)), func() {
		game.events <- GameEvent{Type: RESET_PLAYER_VELOCITY, Data: player.Position}
	})

	return player, nil
}

var gameStore = map[string]*Game{}

func CreateGame(info NewGameInfo) *Game {
	var balls [1](*Ball)
	centrePosition := [2]float32{
		float32(GAME_WIDTH / 2),
		float32(GAME_HEIGHT / 2),
	}

	velocity := [2]float32{
		float32(1),
		float32(1),
	}

	firstBall := Ball{
		CentrePosition: &centrePosition,
		Velocity:       &velocity,
		Radius:         float32(DEFAULT_BALL_RADIUS),
		IsVisible:      true,
	}

	balls[0] = &firstBall

	game := &Game{
		Id:                uuid.NewString(),
		IsActive:          false,
		Width:             GAME_WIDTH,
		Height:            GAME_HEIGHT,
		NPlayers:          info.NPlayers,
		CreatorName:       info.CreatorName,
		NPlayersConnected: 0,
		events:            make(chan GameEvent),
		StopStatusUpdates: make(chan bool),
		StopGame:          make(chan bool),
		Balls:             &balls,
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

func allPlayersAreReady(game *Game) bool {
	if !game.TopPlayer.IsActive {
		return false
	}

	if game.BottomPlayer == nil || !game.BottomPlayer.IsActive {
		return false
	}

	if game.NPlayers == 2 {
		return true
	}

	if game.LeftPlayer == nil || !game.LeftPlayer.IsActive {
		return false
	}

	if game.NPlayers == 3 {
		return true
	}

	return game.RightPlayer != nil && game.RightPlayer.IsActive
}

func (game *Game) listenForEvents() {
listener:
	for {
		event := <-game.events

		switch event.Type {
		case CLEAN_UP_CONNECTION:
			cleanUpConnection(game, event.Data.(string))
		case ACTIVATE_PLAYER:
			activatePlayer(game, event.Data.(string))
			if allPlayersAreReady(game) {
				go startGame(game)
			}
		case START_STATUS_UPDATES:
			go startStatusUpdates(game)
		case STOP_LISTENING:
			break listener
		case REGISTER_PLAYER_MOVEMENT:
			registerPlayerMovement(game, event.Data.(RegisterPlayerMovementData))
		case TERMINATE_GAME:
			terminateGame(game)
		case EJECT_PLAYER:
			ejectPlayer(game, event.Data.(string))
		case RESET_PLAYER_VELOCITY:
			player := getPlayer(game, event.Data.(string))
			player.Velocity = 0
		}
	}
}

func attachDataChannelHandlers(game *Game, player *Player) {
	dataChannel := player.DataChannel
	label := dataChannel.Label()

	// Register channel opening handling
	dataChannel.OnOpen(handleDataChannelOpen(game, player))

	// Register text message handling
	dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
		game.events <- GameEvent{Type: REGISTER_PLAYER_MOVEMENT, Data: RegisterPlayerMovementData{
			Player:  player,
			Message: message.Data,
		}}
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

func RegisterPeerConnection(gameId string, peerConnection *webrtc.PeerConnection, playerName string) error {
	game, ok := GetGame(gameId)
	if !ok {
		return errors.New("no such game")
	}

	player, err := game.FindPlayerToAssign()

	if err != nil {
		return err
	}

	player.PeerConnection = peerConnection
	player.Name = playerName
	attachPeerConnectionHandlers(game, player, peerConnection)
	return nil
}
