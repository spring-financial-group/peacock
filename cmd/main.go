package main

import (
	"os"

	"spring-financial-group/peacock/cmd/app"
)

func main() {
	if err := app.Run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
