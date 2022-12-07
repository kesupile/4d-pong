package main

import (
	"log"
	"net/http"
	"server/internal"
)

func main() {
	gameDB := internal.CreateDB()

	log.Println(gameDB)

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.ListenAndServe(":3000", nil)
}
