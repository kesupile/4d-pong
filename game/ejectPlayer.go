package games

func ejectPlayer(game *Game, side string) {
	player := maybeGetActivePlayer(game, side)
	if player == nil || !player.IsActive {
		return
	}

	player.IsActive = false
	player.IsEjected = true

	tiggerTerminationIfRequired(game)
}
