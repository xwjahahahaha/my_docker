package container

import (
	"os"
	"syscall"
	"xwj/mydocker/log"
)

// RunContainerInitProcess
// @Description: 容器内部执行的函数
// @param cmd
// @param args
// @return error
func RunContainerInitProcess(cmd string, args []string) error {
	log.Log.Infof("command %s", cmd)
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
	argsv := []string{cmd}
	if err := syscall.Exec(cmd, argsv, os.Environ()); err != nil {
		log.Log.WithField("method", "syscall.Exec").Error(err)
		return err
	}
	return nil
}