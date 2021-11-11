package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"xwj/mydocker/log"
)


func RunContainerInitProcess() error {
	// 从管道中读取用户的所有命令
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf(" Run container get user command error, user cmd is nil.")
	}
	// 首先设置/proc为私有模式，防止影响外部/proc
	if err := syscall.Mount("", "/proc", "proc", syscall.MS_REC | syscall.MS_PRIVATE, ""); err != nil {
		log.Log.WithField("method", "syscall.Mount").Error(err)
		return err
	}
	// 挂载/proc文件系统
	// 设置挂载点的flag
	defaultMountFlags :=  syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Log.WithField("method", "syscall.Mount").Error(err)
		return err
	}
	// 寻找在系统PATH下该命令的绝对路径  cmdArray[0]就是命令，后面的都是flag或其他参数
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		return fmt.Errorf(" Exec look path error : %v", err)
	}
	log.Log.Infof("Find path %s", path)
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		log.Log.WithField("method", "syscall.Exec").Error(err)
		return err
	}
	return nil
}

// readUserCommand
// @Description: 读取用户命令
// @return []string
func readUserCommand() []string {
	// 读取文件描述为3的文件, 也就是传递过来的管道的读取端
	pipeReader := os.NewFile(uintptr(3), "pipe")
	// 读取管道中的所有数据
	cmds, err := ioutil.ReadAll(pipeReader)
	if err != nil {
		log.LogErrorFrom("readUserCommand", "ioutil.ReadAll", err)
		return nil
	}
	cmdStrs := string(cmds)
	// 按空格分割命令
	return strings.Split(cmdStrs, " ")
}