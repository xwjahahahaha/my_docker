package container

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
	"xwj/mydocker/log"
)

// NewParentProcess
// @Description: 创建新的命令进程(并未执行)
// @param tty
// @return *exec.Cmd
// @return *os.File   管道写入端
func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {
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
	rootUrl := "./"
	mntUrl := "./mnt"
	imageName := "busybox"
	NewWorkSpace(rootUrl, imageName, mntUrl)
	cmd.Dir = mntUrl 			// 设置进程启动的路径
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

// NewWorkSpace
// @Description: 创建新的文件工作空间
// @param rootURL
// @param mntURL
func NewWorkSpace(rootURL, imageName, mntURL string) {
	CreateReadOnlyLayer(rootURL, imageName)      // 创建init只读层
	CreateWriteLayer(rootURL)                    // 创建读写层
	CreateMountPoint(rootURL, imageName, mntURL) // 创建mnt文件夹并挂载
}

// CreateReadOnlyLayer
// @Description: 通过镜像的压缩包解压并创建镜像文件夹作为只读层
// @param rootURL
// @param imageName
func CreateReadOnlyLayer(rootURL, imageName string) {
	imageName = strings.Trim(imageName, "/")
	imageDir := rootURL + imageName + "/"
	imageTarPath := rootURL + imageName + ".tar"
	if has, err := dirOrFileExist(imageTarPath); err == nil && !has {
		log.Log.Errorf(" Target image tar file not exist!")
		return
	}
	if has, err := dirOrFileExist(imageDir); err == nil && !has {
		// 创建文件夹
		if err := os.Mkdir(imageDir, 0777); err != nil {
			log.LogErrorFrom("createReadOnlyLayer", "Mkdir", err)
		}
	}
	if _, err := exec.Command("tar", "-xvf", imageTarPath, "-C", imageDir).CombinedOutput(); err != nil {
		log.LogErrorFrom("createReadOnlyLayer", "tar", err)
	}
}

// CreateWriteLayer
// @Description: 创建读写层
// @param rootURL
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if has, err := dirOrFileExist(writeURL); err == nil && has {
		log.Log.Info("Write layer dir already exist. Delete and create new one.")
		// 如果存在则先删除掉之前的
		DeleteWriteLayer(rootURL)
	}
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.LogErrorFrom("createWriteLayer", "Mkdir", err)
	}
}

// CreateMountPoint
// @Description: 挂载到容器目录mnt
// @param rootURL
// @param imageName
// @param mntURL
func CreateMountPoint(rootURL, imageName, mntURL string) {
	if has, err := dirOrFileExist(mntURL); err == nil && has {
		log.Log.Info("mnt dir already exist. Delete and create new one.")
		DeleteMountPoint(mntURL)
	}
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.LogErrorFrom("CreateMountPoint", "Mkdir", err)
	}
	// 将读写层目录与镜像只读层目录mount到mnt目录下
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + imageName
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "myDockerMnt", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.LogErrorFrom("createMountPoint", "mount", err)
	}
}

// DeleteWorkSpace
// @Description: 当容器删除时一起删除工作空间
// @param rootURL
// @param mntURL
func DeleteWorkSpace(rootURL, mntURL string) {
	// 镜像层的目录不需要删除
	DeleteMountPoint(mntURL)
	DeleteWriteLayer(rootURL)
}

// DeleteMountPoint
// @Description: 取消挂载点并删除mnt目录
// @param mntURL
func DeleteMountPoint(mntURL string) {
	// 取消mnt目录的挂载
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.LogErrorFrom("deleteMountPoint", "umount", err)
	}
	// 删除mnt目录
	if err := os.RemoveAll(mntURL); err != nil {
		log.LogErrorFrom("deleteMountPoint", "remove", err)
	}
}

// DeleteWriteLayer
// @Description: 删除读写层目录
// @param rootURL
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.LogErrorFrom("deleteWriteLayer", "remove", err)
	}
}
