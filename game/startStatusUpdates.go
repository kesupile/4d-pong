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

func getGamePositions(game *Game) []byte {
	positions := makePositionsArray()

	positions[0] = byte(0)

	if game.TopPlayer != nil && game.TopPlayer.IsActive {
		positions[1] = byte(1)
		positions[3] = byte(game.TopPlayer.Coordinates[0])
		positions[4] = byte(game.TopPlayer.Coordinates[1])
		positions[5] = byte(game.TopPlayer.MagX)
		positions[6] = byte(game.TopPlayer.MagY)
	}

	if game.BottomPlayer != nil && game.BottomPlayer.IsActive {
		positions[7] = byte(1)
		positions[9] = byte(game.BottomPlayer.Coordinates[0])
		positions[10] = byte(game.BottomPlayer.Coordinates[1])
		positions[11] = byte(game.BottomPlayer.MagX)
		positions[12] = byte(game.BottomPlayer.MagY)
	}

	if game.LeftPlayer != nil && game.LeftPlayer.IsActive {
		positions[13] = byte(1)
		positions[15] = byte(game.LeftPlayer.Coordinates[0])
		positions[16] = byte(game.LeftPlayer.Coordinates[1])
		positions[17] = byte(game.LeftPlayer.MagX)
		positions[18] = byte(game.LeftPlayer.MagY)
	}

	if game.RightPlayer != nil && game.RightPlayer.IsActive {
		positions[19] = byte(1)
		positions[21] = byte(game.RightPlayer.Coordinates[0])
		positions[22] = byte(game.RightPlayer.Coordinates[1])
		positions[23] = byte(game.RightPlayer.MagX)
		positions[24] = byte(game.RightPlayer.MagY)
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
		case <-game.stopStatusUpdates:
			game.StatusUpdatesActive = false
			ticker.Stop()
		case <-ticker.C:
			sendGameStatusToPlayers(game)
		}
	}
}
