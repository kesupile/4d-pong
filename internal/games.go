package internal

type Player struct {
	Position string
}

type Game struct {
	Id           string
	Active       string
	Width        int
	Height       int
	TopPlayer    Player
	BottomPlayer Player
	LeftPlayer   Player
	RightPlayer  Player
}

func CreateDB() map[string]Game {
	mapDB := make(map[string]Game)
	return mapDB
}
