package container

import (
	"os"
	"os/exec"
	"syscall"
	"xwj/mydocker/log"
)



// NewParentProcess
// @Description: 创建新的命令进程(并未执行)
// @param tty
// @param command
// @return *exec.Cmd
func NewParentProcess(tty bool, command string) *exec.Cmd {
	// 调用init初始化一些进程的环境和资源
	args := []string{"init", command}
	// 设置/proc/self/exe的命令就是调用自己
	cmd := exec.Command("/proc/self/exe", args...)
	// 使用Clone参数设置隔离环境
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 如果设置了交互，就把输出都导入到标准输入输出中
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

// Run
// @Description: 执行命令
// @param tty
// @param cmd
func Run(tty bool, cmd string){
	parent := NewParentProcess(tty, cmd)
	// 执行命令但是并不等待其结束
	// 执行后会clone出一个namespace隔离的进程，然后在子进程中调用/proc/self/exe即自己，
	// 发送init参数调用init方法初始化一些资源
	if err := parent.Start(); err != nil {
		log.Log.Error(err)
	}
	// 等待结束
	if err := parent.Wait(); err != nil {
		log.Log.Error(err)
	}
	os.Exit(1)
}