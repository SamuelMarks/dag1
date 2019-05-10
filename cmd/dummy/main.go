package main

import (
	cmd "github.com/SamuelMarks/dag1/cmd/dummy/commands"
	_ "net/http/pprof"
	"os"
)

func main() {
	rootCmd := cmd.RootCmd
	//Do not print usage when error occurs
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
