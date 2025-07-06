package main

import (
	"fmt"
	"go-video-streamer/handler"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	const port = 8080

	router := mux.NewRouter()
	router.HandleFunc("/", handler.Index).Methods("GET")
	router.HandleFunc("/video", handler.VideoHandler).Methods("GET")

	fmt.Printf("Starting server on Port : %v\nVisit http://localhost:8080 to see the result.", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), router))
}
