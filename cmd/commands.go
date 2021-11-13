package cmd

import (
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
	CgroupName       = "myDockerTestCgroup"         // 新建的cgroup的名称
	ConfigPath       string                         //配置文件路径
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
	Run: func(cmd *cobra.Command, args []string) {
		// 获取交互flag值与command, 启动容器
		container.Run(tty, strings.Split(args[0], " "), ResourceLimitCfg, CgroupName)
	},
}
