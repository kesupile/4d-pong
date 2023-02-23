package games

import (
	"time"
)

func sendMessageForPlayer(player *Player, positions []byte) {
	index := 2

	switch player.Position {
	case "bottom":
		index = 8
	case "left":
		index = 14
	case "right":
		index = 20
	}

	positions[index] = byte(1)
	player.DataChannel.Send(positions)
}

func makePositionsArray() []byte {
	return make([]byte, 29)
}

func normaliseBallPosition(ball *Ball) (int, int, int) {
	radius := int(ball.Radius)
	x := int(ball.CentrePosition[0]) - radius
	y := int(ball.CentrePosition[1]) - radius
	return radius, x, y
}

func updateBallPositions(game *Game, positions []byte) {
	ball := game.Balls[0]
	if !ball.IsVisible {
		return
	}

	radius, x, y := normaliseBallPosition(ball)

	positions[25] = byte(1)
	positions[26] = byte(radius)
	positions[27] = byte(x)
	positions[28] = byte(y)
}

func setPlayerPosition(player *Player, positions []byte, startIndex int) {
	positions[startIndex] = byte(1)
	positions[startIndex+2] = byte(player.Coordinates[0])
	positions[startIndex+3] = byte(player.Coordinates[1])
	positions[startIndex+4] = byte(player.MagX)
	positions[startIndex+5] = byte(player.MagY)
}

func getGamePositions(game *Game) []byte {
	positions := makePositionsArray()

	positions[0] = byte(0)

	if playerIsActive(game, "top") {
		setPlayerPosition(game.TopPlayer, positions, 1)
	}

	if playerIsActive(game, "bottom") {
		setPlayerPosition(game.BottomPlayer, positions, 7)
	}

	if playerIsActive(game, "left") {
		setPlayerPosition(game.LeftPlayer, positions, 13)
	}

	if playerIsActive(game, "right") {
		setPlayerPosition(game.RightPlayer, positions, 19)
	}

	updateBallPositions(game, positions)

	return positions
}

func sendGameStatusToPlayers(game *Game) {
	positions := getGamePositions(game)

	var generatePositionsCopy = func() []byte {
		destination := makePositionsArray()
		copy(destination, positions)
		return destination
	}

	if playerIsActive(game, "top") {
		sendMessageForPlayer(game.TopPlayer, generatePositionsCopy())
	}

	if playerIsActive(game, "bottom") {
		sendMessageForPlayer(game.BottomPlayer, generatePositionsCopy())
	}

	if playerIsActive(game, "left") {
		sendMessageForPlayer(game.LeftPlayer, generatePositionsCopy())
	}

	if playerIsActive(game, "right") {
		sendMessageForPlayer(game.RightPlayer, generatePositionsCopy())
	}
}

func startStatusUpdates(game *Game) {

	if game.StatusUpdatesActive {
		return
	}

	ticker := time.NewTicker(17 * time.Millisecond)

	game.StatusUpdatesActive = true

	for {
		select {
		case <-game.StopStatusUpdates:
			game.StatusUpdatesActive = false
			ticker.Stop()
		case <-ticker.C:
			sendGameStatusToPlayers(game)
		}
	}
}
