package games

import "log"

func cleanUpConnection(game *Game, position string) {
	log.Println("Cleaning up connection", position)

	var player *Player
	switch position {
	case "top":
		player = game.TopPlayer
		game.TopPlayer = nil
	case "bottom":
		player = game.BottomPlayer
		game.BottomPlayer = nil
	case "left":
		player = game.LeftPlayer
		game.LeftPlayer = nil
	case "right":
		player = game.RightPlayer
		game.RightPlayer = nil
	}

	game.NPlayersConnected = game.NPlayersConnected - 1
	dataChannel := player.DataChannel

	if dataChannel != nil {
		dataChannel.Close()
	}

	player.PeerConnection.Close()
}
