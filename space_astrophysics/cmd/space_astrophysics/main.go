package main

import (
	"log"

	"space_astrophysics/internal/api"
)

func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated")
}
