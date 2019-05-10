package commands

import (
	"github.com/spf13/cobra"
)

var (
	config = NewDefaultCLIConfig()
)

//RootCmd is the root command for DAG1
var RootCmd = &cobra.Command{
	Use:              "dag1",
	Short:            "dag1 consensus",
	TraverseChildren: true,
}
