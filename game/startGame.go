package games

import "time"

func calculateGameStatus(game *Game) {}

func startGame(game *Game) {

	ticker := time.NewTicker(17 * time.Millisecond)

	game.Active = true

	for {
		select {
		case <-game.StopGame:
			game.Active = false
			ticker.Stop()
		case <-ticker.C:
			calculateGameStatus(game)
		}
	}
}
