package games

func terminateGame(game *Game) {
	game.StopGame <- true

	// TODO: Notify winning player wag1

	// TODO: Close all connections
}
