package main

import (
	"log"
	"net/http"
	"server/internal"
)

func main() {
	gameDB := internal.CreateDB()

	log.Println(gameDB)

	go internal.Connect()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/api/session-start", internal.HandleSessionStart)
	http.ListenAndServe(":3000", nil)
}
