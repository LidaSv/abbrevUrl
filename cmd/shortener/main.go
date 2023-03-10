package main

import (
	"abbrevUrl/internal/server"
	"log"
)

func main() {
	err := server.AddServer()
	if err != nil {
		log.Fatal(err)
	}
}
