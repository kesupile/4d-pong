package games

func cleanUpConnection(game *Game, side string) {
	player := getPlayer(game, side)
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
