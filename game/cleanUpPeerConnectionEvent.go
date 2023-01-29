package games

func cleanUpConnection(game *Game, position string) {
	var player *Player
	switch position {
	case "top":
		player = game.TopPlayer
	case "bottom":
		player = game.BottomPlayer
	case "left":
		player = game.LeftPlayer
	case "right":
		player = game.RightPlayer
	}

	player.IsActive = false
	player.IsEjected = true
	game.NPlayersConnected = game.NPlayersConnected - 1
	dataChannel := player.DataChannel

	if dataChannel != nil {
		dataChannel.Close()
	}

	player.PeerConnection.Close()

	tiggerTerminationIfRequired(game)
}
