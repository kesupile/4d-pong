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
	x := int(ball.CentrePosition[0]) - radius/2
	y := int(ball.CentrePosition[1]) - radius/2
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

	if game.TopPlayer != nil && game.TopPlayer.IsActive {
		setPlayerPosition(game.TopPlayer, positions, 1)
	}

	if game.BottomPlayer != nil && game.BottomPlayer.IsActive {
		setPlayerPosition(game.BottomPlayer, positions, 7)
	}

	if game.LeftPlayer != nil && game.LeftPlayer.IsActive {
		setPlayerPosition(game.TopPlayer, positions, 13)
	}

	if game.RightPlayer != nil && game.RightPlayer.IsActive {
		setPlayerPosition(game.TopPlayer, positions, 19)
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

	if game.TopPlayer != nil && game.TopPlayer.IsActive {
		sendMessageForPlayer(game.TopPlayer, generatePositionsCopy())
	}

	if game.BottomPlayer != nil && game.BottomPlayer.IsActive {
		sendMessageForPlayer(game.BottomPlayer, generatePositionsCopy())
	}

	if game.LeftPlayer != nil && game.LeftPlayer.IsActive {
		sendMessageForPlayer(game.LeftPlayer, generatePositionsCopy())
	}

	if game.RightPlayer != nil && game.RightPlayer.IsActive {
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
