package cmd

import (
	"github.com/spf13/cobra"
	"xwj/mydocker/container"
)

const (
	initUsage = `Init container process run user's process in container.Do not call it outside.`
	runUsage = `Create a container with namespace and cgroups limit: myDocker run -t [command]`
)

var (
	tty bool					// 是否交互式执行
)

var initDocker = &cobra.Command{
	Use:   "init [command]",
	Short: initUsage,
	Long:  initUsage,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取传递的command参数，执行容器的初始化操作
		return container.RunContainerInitProcess(args[0], nil)
	},
}

var runDocker = &cobra.Command{
	Use:   "run [command]",
	Short: runUsage,
	Long:  runUsage,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取交互flag值与command, 启动容器
		container.Run(tty, args[0])
	},
}