package main

import (
	"awesomeProject/api"
	"awesomeProject/configs"
)

func main() {
	config := configs.Configs{}
	config.Load()

	a := api.Api{}
	// You need to set your Username and Password here
	a.Initialize(config)

	a.Run(":80")
}
