package games

func cleanUpConnection(game *Game, position string) {
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

	player.IsActive = false
	game.NPlayersConnected = game.NPlayersConnected - 1
	dataChannel := player.DataChannel

	if dataChannel != nil {
		dataChannel.Close()
	}

	player.PeerConnection.Close()
}
