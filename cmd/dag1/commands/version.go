package commands

import (
	"fmt"

	"github.com/SamuelMarks/dag1/src/version"
	"github.com/spf13/cobra"
)

// VersionCmd displays the version of dag1 being used
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}
