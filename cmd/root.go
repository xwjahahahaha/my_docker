package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)


var rootCMD = &cobra.Command{
	Use:   "myDocker",
	Long: `myDocker is a simple container runtime implementation.`,
}

func Execute() {
	if err := rootCMD.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}