package container

import (
	"fmt"
	"os"
	"strings"
	"xwj/mydocker/cgroups"
	"xwj/mydocker/cgroups/subsystems"
	"xwj/mydocker/log"
)

// Run
// @Description: 执行命令
// @param tty
// @param cmdArray
// @param res
// @param cgroupName
// @param volume
func Run(tty bool, cmdArray []string, res *subsystems.ResourceConfig, cgroupName string, volume string){
	// 获取到管道写端
	parent, pipeWriter := NewParentProcess(tty, volume)
	if parent == nil {
		log.LogErrorFrom("Run", "NewParentProcess", fmt.Errorf(" parent process is nil"))
		return
	}
	// 执行命令但是并不等待其结束
	// 执行后会clone出一个namespace隔离的进程，然后在子进程中调用/proc/self/exe即自己，
	// 发送init参数调用init方法初始化一些资源
	if err := parent.Start(); err != nil {
		log.Log.Error(err)
	}
	// 发送用户的命令
	sendUserCommand(cmdArray, pipeWriter)
	// 创建cgroup manager并通过调用set和apply设置资源限制并在容器上生效
	cgroupManager := cgroups.NewCgroupManager(cgroupName)
	// 设置资源限制
	cgroupManager.Set(res)
	// 将容器进程加入到各个子系统中
	cgroupManager.Apply(parent.Process.Pid)
	// 等待结束
	if err := parent.Wait(); err != nil {
		log.Log.Error(err)
	}
	cgroupManager.Destroy()
	// 删除设置的AUFS工作目录
	rootUrl := "./"
	mntUrl := "./mnt"
	DeleteWorkSpace(rootUrl, mntUrl, volume)
	os.Exit(1)
}

// sendUserCommand
// @Description: 想子进程管道中发送命令
// @param cmdArray
// @param pipeWriter
func sendUserCommand(cmdArray []string, pipeWriter *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Log.Infof("First execute cmd is %s", command)
	if _, err := pipeWriter.WriteString(command); err != nil {
		log.LogErrorFrom("sendUserCommand", "WriteString", err)
		return
	}
	err := pipeWriter.Close()
	if err != nil {
		log.LogErrorFrom("sendUserCommand", "Close", err)
		return
	}
}