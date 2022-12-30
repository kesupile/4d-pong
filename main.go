package main

import (
	"net/http"

	"server/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// HTML
	r.Handle("/", handlers.HandleStatic("index.html"))
	r.Handle("/game/{gameId:[\\w|-]+}", handlers.HandleValidatedStatic("game.html", handlers.ValidateGameId))

	// JS
	r.Handle("/index.js", handlers.HandleStatic("index.js"))
	r.Handle("/game/game.js", handlers.HandleStatic("game.js"))

	r.Get(
		"/api/game/{gameId:[\\w|-]+}/status",
		handlers.HandleValidatedRestEndpoint(handlers.ValidateGameId, handlers.HandleGameStatusGET),
	)

	r.Post(
		"/api/game/{gameId:[\\w|-]+}/join",
		handlers.HandleValidatedRestEndpoint(handlers.ValidateGameId, handlers.HandleGameJoinPOST),
	)

	r.Post("/api/new-game", http.HandlerFunc(handlers.HandleNewGamePOST))

	http.ListenAndServe(":4000", r)
}
