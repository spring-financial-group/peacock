package app

import (
	"spring-financial-group/mqube-go-cli-barebones/pkg/cmd"
)

// Run runs the command, if args are not nil they will be set on the command
func Run(args []string) error {
	rootCmd := cmd.Main()
	if args != nil {
		args = args[1:]
		rootCmd.SetArgs(args)
	}
	return rootCmd.Execute()
}
