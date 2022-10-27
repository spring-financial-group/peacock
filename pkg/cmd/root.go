package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/peacock/pkg/cmd/run"
	"github.com/spring-financial-group/peacock/pkg/cmd/version"
	"github.com/spring-financial-group/peacock/pkg/rootcmd"
	"github.com/spring-financial-group/peacock/pkg/utils"
)

// Main creates the new command
func Main() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootcmd.TopLevelCommand,
		Short: "a CICD CLI tool for notifying teams about new releases",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			utils.CheckErr(err)
		},
	}
	// Initialise logger
	var isVerbose bool
	cmd.PersistentFlags().BoolVarP(&isVerbose, "verbose", "v", false, "verbose output")
	utils.InitLogger(isVerbose)

	cmd.AddCommand(run.NewCmdRun())
	cmd.AddCommand(utils.SplitCommand(version.NewCmdVersion()))
	return cmd
}
