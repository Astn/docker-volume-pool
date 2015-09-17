package main

import (
	"fmt"
	"html"
	"net/http"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Echo, %q",
		html.EscapeString(r.URL.Path))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", echoHandler)
	mux.HandleFunc("/VolumeDriver.Create", echoHandler)
	mux.HandleFunc("/VolumeDriver.Mount", echoHandler)
	mux.HandleFunc("/VolumeDriver.Path", echoHandler)
	mux.HandleFunc("/VolumeDriver.Unmount", echoHandler)

	fmt.Println("starting up..")
	fmt.Println(http.ListenAndServe(":8081", mux))
	fmt.Println("shutting down..")
}
