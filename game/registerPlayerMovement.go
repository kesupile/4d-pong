package games

type RegisterPlayerMovementData struct {
	Player  *Player
	Message []byte
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func updateHorizontalPlayerPosition(game *Game, player *Player, message []byte) {
	minPlayerPosition := 0
	maxPlayerPosition := game.Width - player.MagX

	movement := int(message[1])
	currentX := player.Coordinates[0]
	var nextXPosition int

	if movement == 0 {
		// Left
		nextXPosition = max(minPlayerPosition, currentX-MOVEMENT_SCALE_FACTOR)
	} else {
		// Right
		nextXPosition = min(maxPlayerPosition, currentX+MOVEMENT_SCALE_FACTOR)
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
		nextYPosition = max(minPlayerPosition, currentY-MOVEMENT_SCALE_FACTOR)
	} else {
		// Right
		nextYPosition = min(maxPlayerPosition, currentY+MOVEMENT_SCALE_FACTOR)
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

}
