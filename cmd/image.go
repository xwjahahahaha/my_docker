package cmd

import (
	"github.com/spf13/cobra"
	"xwj/mydocker/container"
)

var commitContainerCMD = &cobra.Command{
	Use:   "commit [container_id] [image_tar_name]",
	Short: "commit a container into image",
	Long:  "commit a container into image",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		container.CommitContainer(args[0], args[1])
	},
}
