package games

func playerIsActive(game *Game, side string) bool {
	if side == "top" {
		return game.TopPlayer != nil && game.TopPlayer.IsActive
	}

	if side == "bottom" {
		return game.BottomPlayer != nil && game.BottomPlayer.IsActive
	}

	if side == "left" {
		return game.LeftPlayer != nil && game.LeftPlayer.IsActive
	}

	return game.RightPlayer != nil && game.RightPlayer.IsActive
}

func maybeGetActivePlayer(game *Game, side string) *Player {
	if playerIsActive(game, side) {
		switch side {
		case "top":
			return game.TopPlayer
		case "bottom":
			return game.BottomPlayer
		case "left":
			return game.LeftPlayer
		default:
			return game.RightPlayer
		}
	}
	return nil
}

func startTermination(gameToEnd *Game) {
	gameToEnd.events <- GameEvent{
		Type: TERMINATE_GAME,
	}
}

func tiggerTerminationIfRequired(game *Game) {
	if !game.IsActive {
		return
	}

	players := []*Player{
		game.TopPlayer,
		game.BottomPlayer,
		game.LeftPlayer,
		game.RightPlayer,
	}

	ejectedCount := 0
	for _, player := range players {
		if player != nil && player.IsEjected {
			ejectedCount += 1
		}
	}

	if ejectedCount >= game.NPlayers-1 {
		go startTermination(game)
	}
}
