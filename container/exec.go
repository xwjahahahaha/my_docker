package container

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"xwj/mydocker/log"
)

const (
	ENV_EXEC_PID = "mydocker_pid"
	ENV_EXEC_CMD = "mydocker_cmd"
)

// ExecContainer
// @Description: 创建子命令运行exec
// @param containerID
// @param commandAry
func ExecContainer(containerID string, commandAry []string)  {
	pid, err := getContainerPidByID(containerID)
	if err != nil {
		return
	}
	cmdStr := strings.Join(commandAry, " ")
	log.Log.Infof("container pid %s", pid)
	log.Log.Infof("command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置环境变量：进程号与执行命令
	if err := os.Setenv(ENV_EXEC_PID, pid); err != nil {
		log.LogErrorFrom("ExecContainer", "Setenv_ENV_EXEC_PID", err)
		return
	}
	if err := os.Setenv(ENV_EXEC_CMD, cmdStr); err != nil {
		log.LogErrorFrom("ExecContainer", "Setenv_ENV_EXEC_CMD", err)
		return
	}

	if err := cmd.Run(); err != nil {
		log.LogErrorFrom("ExecContainer", "Run", err)
		return
	}
}

// getContainerPidByID
// @Description: 根据容器ID获取其PID
// @param containerID
// @return string
// @return error
func getContainerPidByID(containerID string) (string, error)  {
	// 读取容器信息文件
	containerInfoPath := filepath.Join(DefaultInfoLocation, containerID, ConfigName)
	content, err := ioutil.ReadFile(containerInfoPath)
	if err != nil {
		log.LogErrorFrom("getContainerPidByID", "ReadFile", err)
		return "", err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.LogErrorFrom("getContainerPidByID", "Unmarshal", err)
		return "", err
	}
	return containerInfo.Pid, nil
}

// getContainerByID
// @Description: 根据容器ID获取容器信息结构体
// @param containerID
// @return *ContainerInfo
// @return error
func getContainerByID(containerID string) (*ContainerInfo, error)  {
	// 读取容器信息文件
	containerInfoPath := filepath.Join(DefaultInfoLocation, containerID, ConfigName)
	content, err := ioutil.ReadFile(containerInfoPath)
	if err != nil {
		log.LogErrorFrom("getContainerByID", "ReadFile", err)
		return nil, err
	}
	var containerInfo ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.LogErrorFrom("getContainerByID", "Unmarshal", err)
		return nil, err
	}
	return &containerInfo, nil
}

// StopContainer
// @Description: 关闭容器
// @param containerID
func StopContainer(containerID string)  {
	containerInfo, err := getContainerByID(containerID)
	if err != nil {
		log.LogErrorFrom("StopContainer", "getContainerByID", err)
		return
	}
	// 系统调用kill可以发送信号给进程，通过传递syscall.SIGTERM信号，去杀掉容器主进程
	pid, _ := strconv.Atoi(containerInfo.Pid)
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		log.LogErrorFrom("StopContainer", "Kill", err)
		return
	}
	// 修改容器的状态
	containerInfo.Status = STOP
	containerInfo.Pid = " "				// 注意这里要设置一个空格，为了exec判断pid不为空""
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.LogErrorFrom("StopContainer", "Marshal", err)
		return
	}
	configPath := filepath.Join(DefaultInfoLocation, containerID, ConfigName)
	// 写入
	if err := ioutil.WriteFile(configPath, newContentBytes, 0622); err != nil {
		log.LogErrorFrom("StopContainer", "WriteFile", err)
	}
}