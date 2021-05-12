package main

import (
	"awesomeProject/api"
)

func main() {
	a := api.Api{}
	// You need to set your Username and Password here
	a.Initialize()

	a.Run(":80")
}
