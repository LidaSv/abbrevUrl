package main

import (
	"abbrevUrl/internal/app"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", app.Abbrevurl)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
