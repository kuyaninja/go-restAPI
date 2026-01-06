package main

import (
	"log"

	"skeleton-go/app"
)

func main() {
	application := app.NewApplication()

	if err := application.Start(); err != nil {
		log.Fatalf("failed to start application: %v", err)
	}
}
