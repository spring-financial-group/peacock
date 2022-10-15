package utils

import (
	"os"

	"github.com/spf13/cobra"
)

// BinaryName the binary name to use in help docs
var BinaryName string

// TopLevelCommand the top level command name
var TopLevelCommand string

func init() {
	BinaryName = os.Getenv("BINARY_NAME")
	if BinaryName == "" {
		BinaryName = "peacock"
	}
	TopLevelCommand = os.Getenv("TOP_LEVEL_COMMAND")
	if TopLevelCommand == "" {
		TopLevelCommand = "peacock"
	}
}

// SplitCommand helper command to ignore the options object
func SplitCommand(cmd *cobra.Command, _ interface{}) *cobra.Command {
	return cmd
}
