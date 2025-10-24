package utils

import (
	"os"

	"github.com/spf13/cobra"
)

// BinaryName the binary name to use in help docs
var BinaryName = getBinaryName()

// TopLevelCommand the top level command name
var TopLevelCommand = getTopLevelCommand()

func getBinaryName() string {
	name := os.Getenv("BINARY_NAME")
	if name == "" {
		return "peacock"
	}
	return name
}

func getTopLevelCommand() string {
	cmd := os.Getenv("TOP_LEVEL_COMMAND")
	if cmd == "" {
		return "peacock"
	}
	return cmd
}

// SplitCommand helper command to ignore the options object
func SplitCommand(cmd *cobra.Command, _ interface{}) *cobra.Command {
	return cmd
}
