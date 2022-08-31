package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/mqa-helpers/pkg/cobras"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"spring-financial-group/mqube-go-cli-barebones/pkg/cmd/version"
	"spring-financial-group/mqube-go-cli-barebones/pkg/rootcmd"
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
