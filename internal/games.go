package internal

import "github.com/google/uuid"

type Player struct {
	Position string
}

type Game struct {
	Id           string
	Active       bool
	Width        int
	Height       int
	TopPlayer    Player
	BottomPlayer Player
	LeftPlayer   Player
	RightPlayer  Player
}

var gameStore = map[string]Game{}

func CreateGame() Game {
	game := Game{
		Id:     uuid.NewString(),
		Active: false,
		Width:  200,
		Height: 200,
	}

	gameStore[game.Id] = game
	return game
}

func GetGame(id string) (Game, bool) {
	game, ok := gameStore[id]
	return game, ok
}
