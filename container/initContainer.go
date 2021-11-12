package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"xwj/mydocker/log"
)

// RunContainerInitProcess
// @Description: 运行容器的初始化命令进程
// @return error
func RunContainerInitProcess() error {
	// 从管道中读取用户的所有命令
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf(" Run container get user command error, user cmd is nil.")
	}
	// 设置挂载与pivot_root
	setUpMount()
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

// pivotRoot
// @Description: 使用pivot_root更改当前root文件系统
// @param root	指定的新的根目录（一般就是容器的启动目录）
// @return error
func pivotRoot(root string) error {
	// 重新mount新的根目录
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND | syscall.MS_REC, ""); err != nil {
		return fmt.Errorf(" Mount rootfs to itself error: %v", err)
	}
	// 创建临时文件.pivot_root存储old_root
	pivotPath := filepath.Join(root, ".pivot_root")
	// 判断当前目录是否已有该文件夹
	if _ ,err := os.Stat(pivotPath); err == nil {
		// 存在则删除
		if err := os.Remove(pivotPath); err != nil {
			return err
		}
	}
	if err := os.Mkdir(pivotPath, 0777); err != nil {
		return err
	}
	// pivot_root将原根目录挂载到.pivot_root上，然后将root设置为新的根目录文件系统
	if err := syscall.PivotRoot(root, pivotPath); err != nil {
		return fmt.Errorf(" Pivot root err %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf(" Chdir / %v", err)
	}
	// 取消临时文件.pivot_root的挂载并删除它
	pivotPath = filepath.Join("/", ".pivot_root")		// 注意当前已经在根目录下，所以临时文件的目录也改变了
	if err := syscall.Unmount(pivotPath, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf(" Unmount .pivot_root dir %v", err)
	}
	return os.Remove(pivotPath)
}

func setUpMount()  {
	// 首先设置根目录为私有模式，防止影响pivot_root
	if err := syscall.Mount("/", "/", "", syscall.MS_REC | syscall.MS_PRIVATE, ""); err != nil {
		log.LogErrorFrom("setUpMount", "Mount proc", err)
	}
	// 获取当前路径
	pwd, err := os.Getwd()
	if err != nil {
		log.LogErrorFrom("setUpMount", "Getwd", err)
	}
	log.Log.Infof("Current location is %s", pwd)
	// 使用pivot root
	if err := pivotRoot(pwd); err != nil {
		log.LogErrorFrom("setUpMount", "pivotRoot", err)
	}
	// 设置一些挂载
	// 挂载/proc文件系统
	// 设置挂载点的flag
	defaultMountFlags :=  syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.LogErrorFrom("setUpMount", "Mount proc", err)
	}
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID | syscall.MS_STRICTATIME, "mode=755"); err != nil {
		log.LogErrorFrom("setUpMount", "Mount /dev", err)
	}
}