package main

import (
	"net/http"

	"server/internal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	go internal.Connect()

	// HTML
	r.Handle("/", internal.HandleStatic("index.html"))
	r.Handle("/game/{gameId:[\\w|-]+}", internal.HandleValidatedStatic("game.html", internal.ValidateGameId))

	// JS
	r.Handle("/index.js", internal.HandleStatic("index.js"))
	r.Handle("/game/game.js", internal.HandleStatic("game.js"))

	r.Post("/api/session-start", http.HandlerFunc(internal.HandleSessionStart))
	r.Post("/api/new-game", http.HandlerFunc(internal.HandleNewGame))
	http.ListenAndServe(":4000", r)
}
