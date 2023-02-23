package games

import (
	"math"
	"time"
)

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

func calculateTimeToSideCollision(game *Game, ball *Ball, side string) float32 {
	var distance float32
	var speed float32
	switch side {
	case "top":
		distance = ball.CentrePosition[0] - ball.Radius
		if playerIsActive(game, "top") {
			distance -= float32(game.TopPlayer.MagY)
		}
		speed = float32(math.Abs(float64(ball.Velocity[1])))
	case "bottom":
		distance = float32(game.Height) - ball.CentrePosition[1] - ball.Radius
		if playerIsActive(game, "bottom") {
			distance -= float32(game.BottomPlayer.MagY)
		}
		speed = ball.Velocity[1]
	case "left":
		distance = ball.CentrePosition[0] - ball.Radius
		if playerIsActive(game, "left") {
			distance -= float32(game.LeftPlayer.MagX)
		}
		speed = float32(math.Abs(float64(ball.Velocity[0])))
	case "right":
		distance = float32(game.Width) - ball.CentrePosition[0] - ball.Radius
		if playerIsActive(game, "right") {
			distance -= float32(game.RightPlayer.MagX)
		}
		speed = ball.Velocity[0]
	}

	return distance / speed
}

type CollisionDetail struct {
	Side              string
	FramesToCollision float64
}

func calculateAllPotentialCollisionTimes(game *Game, ball *Ball, potentialCollisionSides [2]string) []CollisionDetail {
	var collisionTimes []CollisionDetail
	for _, side := range potentialCollisionSides {
		if side == "" {
			continue
		}

		framesToCollision := calculateTimeToSideCollision(game, ball, side)

		collisionTimes = append(collisionTimes, CollisionDetail{
			Side:              side,
			FramesToCollision: float64(framesToCollision),
		})
	}
	return collisionTimes
}

type FinalCollisionDetails struct {
	Side           string
	FrameFraction  float64
	CheckForPlayer bool
}

func calculateFinalDetail(game *Game, ball *Ball, collisionTime CollisionDetail) FinalCollisionDetails {
	frameFraction := math.Min(1, collisionTime.FramesToCollision)

	return FinalCollisionDetails{
		Side:           collisionTime.Side,
		FrameFraction:  frameFraction,
		CheckForPlayer: playerIsActive(game, collisionTime.Side),
	}
}

func findFirstCollision(game *Game, ball *Ball, allCollisionTimes []CollisionDetail) []FinalCollisionDetails {
	var finalDetails []FinalCollisionDetails
	if len(allCollisionTimes) == 1 {
		return append(
			finalDetails,
			calculateFinalDetail(game, ball, allCollisionTimes[0]),
		)
	}

	if allCollisionTimes[0].FramesToCollision == allCollisionTimes[1].FramesToCollision {
		for _, collisionTime := range allCollisionTimes {
			finalDetails = append(
				finalDetails,
				calculateFinalDetail(game, ball, collisionTime),
			)
		}
		return finalDetails
	}

	if allCollisionTimes[0].FramesToCollision < allCollisionTimes[1].FramesToCollision {
		return append(
			finalDetails,
			calculateFinalDetail(game, ball, allCollisionTimes[0]),
		)
	}

	return append(
		finalDetails,
		calculateFinalDetail(game, ball, allCollisionTimes[1]),
	)
}

func getNextCollisionDetails(game *Game, ball *Ball) []FinalCollisionDetails {
	potentialCollisionSides := getPotentialNextCollisonSides(game, ball)
	allPotentialCollisionTimes := calculateAllPotentialCollisionTimes(game, ball, potentialCollisionSides)
	finalCollisionDetails := findFirstCollision(game, ball, allPotentialCollisionTimes)

	return finalCollisionDetails
}

func updateBallVelocity(ball *Ball, side string) {
	if side == "top" || side == "bottom" {
		ball.Velocity[1] = -1 * ball.Velocity[1]
	} else {
		ball.Velocity[0] = -1 * ball.Velocity[0]
	}
}

func updateBallPosition(ball *Ball) {
	ball.CentrePosition[0] += ball.Velocity[0]
	ball.CentrePosition[1] += ball.Velocity[1]
}

func isBetween(number, lowerBound, upperBound int) bool {
	return lowerBound <= number && number <= upperBound
}

func isPlayerInPosition(game *Game, player *Player, ball *Ball) bool {
	if player.Position == "top" || player.Position == "bottom" {
		x := player.Coordinates[0]
		return isBetween(
			int(ball.CentrePosition[0]),
			x,
			x+PLAYER_WIDTH,
		)
	}

	y := player.Coordinates[1]
	return isBetween(
		int(ball.CentrePosition[1]),
		y,
		y+PLAYER_WIDTH,
	)
}

func startPlayerEjection(game *Game, side string) {
	game.events <- GameEvent{
		Type: EJECT_PLAYER,
		Data: side,
	}
}

func calculateGameStatus(game *Game, finalCollisionDetails []FinalCollisionDetails) {
	if !game.IsActive {
		return
	}

	frameFraction := finalCollisionDetails[0].FrameFraction

	time.Sleep((time.Duration(int(frameFraction)*FRAME_TIME) * time.Millisecond))

	for _, collisionDetails := range finalCollisionDetails {
		if frameFraction < 1 {
			side := collisionDetails.Side
			player := maybeGetActivePlayer(game, side)

		ballCheckBlock:
			for _, ball := range game.Balls {
				if player != nil {
					playerIsInPosition := isPlayerInPosition(game, player, ball)
					if !playerIsInPosition {
						go startPlayerEjection(game, side)
						continue ballCheckBlock
					}
				}

				updateBallVelocity(ball, collisionDetails.Side)
			}
		} else {
			for _, ball := range game.Balls {
				updateBallPosition(ball)
			}
		}
	}

	var nextFinalCollisionDetails []FinalCollisionDetails
	for _, ball := range game.Balls {
		nextFinalCollisionDetails = getNextCollisionDetails(game, ball)
	}

	go calculateGameStatus(game, nextFinalCollisionDetails)
}

func startGame(game *Game) {
	game.StartTime = time.Now().Format("2006-01-02T15:04:05-0700")
	game.IsActive = true

	go calculateGameStatus(game, []FinalCollisionDetails{
		{FrameFraction: 1},
	})

	<-game.StopGame
	game.IsActive = false
}
