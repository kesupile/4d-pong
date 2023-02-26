package games

import "golang.org/x/exp/constraints"

type RegisterPlayerMovementData struct {
	Player  *Player
	Message []byte
}

type Number interface {
	constraints.Ordered
}

func max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func resetPlayerVelocity(player *Player) {}

var positive = "positive"
var negative = "negative"

func updatePlayerVelocity(player *Player, positiveOrNegative string) {
	if positiveOrNegative == positive {
		if player.Velocity < 0 {
			player.Velocity = float32(PLAYER_VELOCITY_INCREMENT)
		} else {
			player.Velocity = min(player.Velocity+float32(PLAYER_VELOCITY_INCREMENT), float32(MAX_PLAYER_VELOCITY))
		}
	} else {
		if player.Velocity > 0 {
			player.Velocity = -float32(PLAYER_VELOCITY_INCREMENT)
		} else {
			player.Velocity = min(player.Velocity-float32(PLAYER_VELOCITY_INCREMENT), float32(-MAX_PLAYER_VELOCITY))
		}
	}
	resetPlayerVelocity(player)
}

func updateHorizontalPlayerPosition(game *Game, player *Player, message []byte) {
	minPlayerPosition := 0
	maxPlayerPosition := game.Width - player.MagX

	movement := int(message[1])
	currentX := player.Coordinates[0]
	var nextXPosition int

	if movement == 0 {
		// Left
		nextXPosition = max(minPlayerPosition, currentX-PLAYER_MOVEMENT_SCALE_FACTOR)
		updatePlayerVelocity(player, negative)
	} else {
		// Right
		nextXPosition = min(maxPlayerPosition, currentX+PLAYER_MOVEMENT_SCALE_FACTOR)
		updatePlayerVelocity(player, positive)
	}

	player.Coordinates[0] = nextXPosition
}

func updateVerticalPlayerPosition(game *Game, player *Player, message []byte) {
	minPlayerPosition := 0
	maxPlayerPosition := game.Height - player.MagY

	movement := int(message[1])
	currentY := player.Coordinates[1]
	var nextYPosition int

	if movement == 0 {
		// Up
		nextYPosition = max(minPlayerPosition, currentY-PLAYER_MOVEMENT_SCALE_FACTOR)
		updatePlayerVelocity(player, negative)
	} else {
		// Down
		nextYPosition = min(maxPlayerPosition, currentY+PLAYER_MOVEMENT_SCALE_FACTOR)
		updatePlayerVelocity(player, positive)
	}

	player.Coordinates[1] = nextYPosition
}

func registerPlayerMovement(game *Game, data RegisterPlayerMovementData) {
	player := data.Player

	if player.Position == "top" || player.Position == "bottom" {
		updateHorizontalPlayerPosition(game, player, data.Message)
	} else {
		updateVerticalPlayerPosition(game, player, data.Message)
	}

	player.ResetVelocity()
}
