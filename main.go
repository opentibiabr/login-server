package main

import (
	"awesomeProject/api"
)

func main() {
	app := api.Api{}
	app.Initialize()
	app.Run(":80")
}
