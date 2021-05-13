package main

import (
	"github.com/opentibiabr/login-server/src/api"
	"log"
)

func main() {
	log.Print("Welcome to OTBR Login Server")
	log.Print("Loading configurations...")

	app := api.Api{}
	app.Initialize()

	app.Configs.Print()
	log.Printf("OTBR Login Server running at port %d!", app.Configs.LoginPort)

	app.Run(":80")
}