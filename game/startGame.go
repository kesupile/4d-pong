package games

import "time"

func findNextCollision(ball *Ball) {

}

func updateBallPosition(ball *Ball) {
	ball.CentrePosition[0] += ball.Velocity[0]
	ball.CentrePosition[1] += ball.Velocity[1]
}

type CollisionDetails struct {
	side           string
	frameFraction  float32
	checkForPlayer bool
}

func getPotentialNextCollisonSides(game *Game, ball *Ball) [2]string {
	var side1 string
	var side2 string
	var sides [2]string
	velocity := ball.Velocity

	if velocity[0] > 0 {
		side1 = "right"
	} else if velocity[0] < 0 {
		side1 = "left"
	}
	sides[0] = side1

	if velocity[1] > 0 {
		side2 = "bottom"
	} else if velocity[1] < 0 {
		side2 = "top"
	}
	sides[1] = side2

	return sides
}

func calculateTimeToSide(game *Game, ball *Ball, side string) {}

func getNextCollisionDetails(game *Game, ball *Ball) CollisionDetails {
	potentialCollisionSides := getPotentialNextCollisonSides(game, ball)

}

func calculateGameStatus(game *Game, frameFraction float32) {
	if !game.Active {
		return
	}

	time.Sleep((time.Duration(frameFraction*17) * time.Millisecond))

	var collisionDetails CollisionDetails
	for _, ball := range game.Balls {
		collisionDetails = getNextCollisionDetails(game, ball)
	}

	go calculateGameStatus(game, 1)
}

func startGame(game *Game) {
	game.Active = true

	go calculateGameStatus(game, 1)

	<-game.StopGame
	game.Active = false
}
