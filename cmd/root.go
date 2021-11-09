package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const usage = `myDocker is a simple container runtime implementation.
The purpose of this project is to learn how docker works and how to write a docker by ourselves
Enjoy it, just for fun.`

var rootCmd = &cobra.Command{
	Use:   "myDocker",
	Short: "myDocker is a simple container runtime implementation",
	Long: usage,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}