package container

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"xwj/mydocker/log"
)

// NewParentProcess
// @Description: 创建新的命令进程(并未执行)
// @param tty
// @return *exec.Cmd
// @return *os.File   管道写入端
func NewParentProcess(tty bool, volume, ImageTarPath, cId string) (*exec.Cmd, *os.File) {
	// 创建匿名管道
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.LogErrorFrom("NewParentProcess", "NewPipe", err)
		return nil, nil
	}
	// 调用init初始化一些进程的环境和资源
	// 设置/proc/self/exe的命令就是调用自己
	cmd := exec.Command("/proc/self/exe", "init")
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
	// 创建新的工作空间
	rootUrl := "/var/lib/mydocker/aufs/" 				  // 根目录
	mntUrl := filepath.Join(rootUrl, "mnt", cId)          // 容器运行空间
	NewWorkSpace(rootUrl, ImageTarPath, mntUrl, volume, cId)
	cmd.Dir = mntUrl 					 // 设置进程启动的路径
	// 在这里传入管道文件读取端的句柄
	// ExtraFiles指定要由新进程继承的其他打开文件。它不包括标准输入、标准输出或标准错误。
	cmd.ExtraFiles = []*os.File{readPipe}
	return cmd, writePipe
}

// NewPipe
// @Description: 创建一个新的匿名管道
// @return *os.File
// @return *os.File
// @return error
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
