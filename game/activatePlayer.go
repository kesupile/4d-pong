package games

func activatePlayer(game *Game, position string) {
	game.NPlayersConnected += 1

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

	player.IsActive = true

	SendPlayerStatusNotification(game)
}
