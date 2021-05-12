package main

import (
	"login-server/api"
)

func main() {
	app := api.Api{}
	app.Initialize()
	app.Run(":80")
}
