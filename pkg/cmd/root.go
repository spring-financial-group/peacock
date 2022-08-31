package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"github.com/spring-financial-group/peacock/pkg/cmd/version"
	"github.com/spring-financial-group/peacock/pkg/rootcmd"
)

// Main creates the new command
func Main() *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootcmd.TopLevelCommand,
		Short: "a CLI template",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Logger().Errorf(err.Error())
			}
		},
	}
	cmd.AddCommand(cobras.SplitCommand(version.NewCmdVersion()))
	return cmd
}
