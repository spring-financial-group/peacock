package version

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spring-financial-group/peacock/pkg/rootcmd"
	"github.com/spring-financial-group/peacock/pkg/utils"

	"github.com/spring-financial-group/peacock/pkg/utils/templates"
)

// Options for triggering
type Options struct {
	Apps []string
	Args []string
	Cmd  *cobra.Command
}

const (
	// TestVersion used in test cases for the current version if no
	// version can be found - such as if the version property is not properly
	// included in the go test flags
	TestVersion = "1.0.0-SNAPSHOT"
)

var (
	createLong = templates.LongDesc(`
		Shows the version of peacock
`)

	createExample = templates.Examples(`
		version
	`)

	Version string
)

// NewCmdVersion creates a new version command
func NewCmdVersion() (*cobra.Command, *Options) {
	o := &Options{}

	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Shows the version of the peacock",
		Long:    createLong,
		Example: fmt.Sprintf(createExample, rootcmd.BinaryName),
		Run: func(cmd *cobra.Command, args []string) {
			o.Cmd = cmd
			o.Args = args
			err := o.Run()
			utils.CheckErr(err)
		},
	}
	o.Cmd = cmd

	return cmd, o
}

func (o *Options) Run() error {
	fmt.Println(GetVersion())
	return nil
}

func GetVersion() string {
	if Version != "" {
		return Version
	}
	return TestVersion
}
