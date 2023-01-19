package games

import (
	"time"
)

func updateBallPosition(ball *Ball) {
	ball.CentrePosition[0] += ball.Velocity[0]
	ball.CentrePosition[1] += ball.Velocity[1]
}

func calculateGameStatus(game *Game) {
	for _, ball := range game.Balls {
		updateBallPosition(ball)
	}
}

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
