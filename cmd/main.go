package main

import (
	"os"

	"spring-financial-group/mqube-go-cli-barebones/cmd/app"
)

func main() {
	if err := app.Run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
