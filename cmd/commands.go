package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
	"xwj/mydocker/cgroups/subsystems"
	"xwj/mydocker/container"
)

const (
	initUsage = `Init container process run user's process in container.Do not call it outside.`
	runUsage  = `Create a container with namespace and cgroups limit: myDocker run -t [command]`
)

var (
	tty              bool                           // 是否交互式执行
	ResourceLimitCfg = &subsystems.ResourceConfig{} // 资源限制配置
	CgroupName       = "myDocker"                   // 新建的cgroup的名称
	Volume           string                         // 数据卷
	Detach           bool                           // 后台运行
	Name             string                         // 容器名称
	ImageTarPath     string                         // 镜像的tar包路径
)

var initDocker = &cobra.Command{
	Use:   "init",
	Short: initUsage,
	Long:  initUsage,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取传递的command参数，执行容器的初始化操作
		return container.RunContainerInitProcess()
	},
}

var runDocker = &cobra.Command{
	Use:   "run [command]",
	Short: runUsage,
	Long:  runUsage,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if tty && Detach {
			// 两个标志不运行同时设置
			return fmt.Errorf(" tty and detach can't both provided.")
		}
		// 生成容器ID
		// 首先生成容器ID
		id := container.RandStringContainerID(10)
		// 获取交互flag值与command, 启动容器
		container.Run(tty, strings.Split(args[0], " "), ResourceLimitCfg, CgroupName, Volume, Name, ImageTarPath, id)
		return nil
	},
}

var commitCommand = &cobra.Command{
	Use:   "commit [image_name]",
	Short: "commit a container into image",
	Long:  "commit a container into image",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		container.CommitContainer(args[0])
	},
}

var listContainers = &cobra.Command{
	Use:   "ps",
	Short: "list all the containers",
	Long:  "list all the containers",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		container.ListAllContainers()
	},
}
