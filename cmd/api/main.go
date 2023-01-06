package main

import (
	"github.com/spring-financial-group/peacock/pkg/server"
	_ "go.uber.org/automaxprocs"
)

// @title Peacock API
// @version 1.0
// @description Service for notifying users of changes to your platform
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @in header
func main() {
	server.Run()
}
