package main

import (
	"os"
	"sport-grid-be/cmd/app"

	logger "github.com/imsab23/platform-be/observability/logging"
)

func main() {
	log, _ := logger.NewLogger("backend-service")

	err := app.RunServer()
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
