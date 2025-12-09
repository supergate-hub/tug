package main

import (
	"log"
	"os"

	"github.com/supergate-hub/tug/cmd/tug/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
