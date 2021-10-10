package main

import (
	"github.com/Aleksluciano/chat386/internal/handlers"
	"log"
	"net/http"
)

func main() {

	mux := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting web server on port ??")

	_ = http.ListenAndServe("", mux)

}
