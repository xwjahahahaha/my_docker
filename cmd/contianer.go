package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
	"xwj/mydocker/container"
	"xwj/mydocker/log"
)

var runContainerCMD = &cobra.Command{
	Use:  "run [command]",
	Long: `Create a container with namespace and cgroups limit: myDocker run -t [command]`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if tty && Detach {
			// 两个标志不运行同时设置
			return fmt.Errorf(" tty and detach can't both provided.")
		}
		// 生成容器ID
		// 首先生成容器ID
		id := container.RandStringContainerID(10)
		log.Log.Infof("Container ID [%s]", id)
		// 获取交互flag值与command, 启动容器
		container.Run(tty, strings.Split(args[0], " "), ResourceLimitCfg, CgroupName, Volume, Name, ImageTarPath, id, EnvSlice, Port, NetWorkName)
		return nil
	},
}

var listContainersCMD = &cobra.Command{
	Use:  "ps",
	Long: "list all the containers",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		container.ListAllContainers()
	},
}

var logContainersCMD = &cobra.Command{
	Use:  "logs [container_id]",
	Long: "print logs of a container",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		container.LogContainer(args[0])
	},
}

var stopContainerCMD = &cobra.Command{
	Use:   "stop [container_id]",
	Short: "stop a container",
	Long:  "stop a container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		container.StopContainer(args[0])
	},
}

var removeContainerCMD = &cobra.Command{
	Use:   "rm [container_id]",
	Short: "remove a container",
	Long:  "remove a container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		container.RemoveContainer(args[0])
	},
}
