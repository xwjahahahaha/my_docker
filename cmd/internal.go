package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"strings"
	"xwj/mydocker/container"
	"xwj/mydocker/log"
	"xwj/mydocker/namespace"
)

const EnvExecPid = "mydocker_pid"

var initContainerCMD = &cobra.Command{
	Use:   "init",
	Long:  `Init container process run user's process in container.Do not call it outside.`,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取传递的command参数，执行容器的初始化操作
		return container.RunContainerInitProcess()
	},
}

var execContainerCMD = &cobra.Command{
	Use:  "exec [container_id] [command]",
	Long: "Exec a command into container",
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv(EnvExecPid) != "" {
			// 第二次调用的时候执行
			log.Log.Infof("pid callback pid %s", os.Getenv(EnvExecPid))
			// 调用namespace包自动调用C代码setns进入容器空间
			namespace.EnterNamespace()
			return
		}
		if len(args) < 2 {
			log.Log.Errorf("Missing container name or command.")
			return
		}
		cid, commandAry := args[0], strings.Split(args[1], " ")
		// 设置环境变量
		container.ExecContainer(cid, commandAry)
	},
}


