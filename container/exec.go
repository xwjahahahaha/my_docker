package container

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"xwj/mydocker/log"
)

const (
	ENV_EXEC_PID = "mydocker_pid"
	ENV_EXEC_CMD = "mydocker_cmd"
)

func ExecContainer(containerID string, commandAry []string)  {
	pid, err := getContainerPidByID(containerID)
	if err != nil {
		return
	}
	cmdStr := strings.Join(commandAry, " ")
	log.Log.Infof("container pid %s", pid)
	log.Log.Infof("command %s", cmdStr)
	// 执行我们自己创建的命令exec
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
	// 执行
	if err := cmd.Run(); err != nil {
		log.Log.Errorf("Exec container %s error %v", containerID, err)
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