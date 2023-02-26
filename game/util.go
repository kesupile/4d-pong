package games

import (
	"sync"
	"time"
)

func getPlayer(game *Game, side string) *Player {
	switch side {
	case "top":
		return game.TopPlayer
	case "bottom":
		return game.BottomPlayer
	case "left":
		return game.LeftPlayer
	default:
		return game.RightPlayer
	}
}

func playerIsActive(game *Game, side string) bool {
	player := getPlayer(game, side)
	return player != nil && player.IsActive
}

func maybeGetActivePlayer(game *Game, side string) *Player {
	if playerIsActive(game, side) {
		return getPlayer(game, side)
	}
	return nil
}

func startTermination(gameToEnd *Game) {
	gameToEnd.events <- GameEvent{
		Type: TERMINATE_GAME,
	}
}

func tiggerTerminationIfRequired(game *Game) {
	if !game.IsActive {
		return
	}

	players := []*Player{
		game.TopPlayer,
		game.BottomPlayer,
		game.LeftPlayer,
		game.RightPlayer,
	}

	ejectedCount := 0
	for _, player := range players {
		if player != nil && player.IsEjected {
			ejectedCount += 1
		}
	}

	if ejectedCount >= game.NPlayers-1 {
		go startTermination(game)
	}
}

type Debounce struct {
	mu       sync.Mutex
	Duration time.Duration
	FnTimer  *time.Timer
}

func (debounce *Debounce) Reset(fn func()) {
	debounce.mu.Lock()
	defer debounce.mu.Unlock()

	if debounce.FnTimer != nil {
		debounce.FnTimer.Stop()
	}

	debounce.FnTimer = time.AfterFunc(debounce.Duration, fn)
}

func WithDebounce(duration time.Duration, timefn func()) func() {
	debounce := Debounce{Duration: duration}
	return func() {
		debounce.Reset(timefn)
	}
}
