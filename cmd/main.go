package main

import (
	"github.com/spring-financial-group/peacock/cmd/cli"
	"os"
)

func main() {
	if err := cli.Run(nil); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
