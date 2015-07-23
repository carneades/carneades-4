package main

import (
	"github.com/carneades/carneades-4/internal/web"
)

const helpServer = `To Do`

func webServerCmd() {
	// webFlags := flag.NewFlagSet("web", flag.ContinueOnError)
	web.CarneadesServer()
}
