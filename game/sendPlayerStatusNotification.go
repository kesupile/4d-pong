package games

func SendPlayerStatusNotification(game *Game) {
	for _, player := range GetAllActivePlayers(game) {
		data := make([]byte, 1)
		data[0] = byte(1)
		player.DataChannel.Send(data)
	}
}
