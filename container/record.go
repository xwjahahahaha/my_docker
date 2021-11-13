package container

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"xwj/mydocker/log"
)

const (
	IDLen = 10
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"created_time"`
	Status      string `json:"status"`
}

var (
	RUNNING             = "running"
	STOP                = "stopped"
	EXIT                = "exited"
	DefaultInfoLocation = "/var/run/mydocker/"
	ConfigName          = "containerInfo.json"
)

// randStringContainerID
// @Description: 容器ID随机生成器
// @param n
// @return string
func randStringContainerID(n int) string {
	if n < 0 || n > 32 {
		n = 32
	}
	// 这里就采用对时间戳取hash的方法实现容器的随机ID生成
	hashBytes := sha256.Sum256([]byte(strconv.Itoa(int(time.Now().UnixNano()))))
	return fmt.Sprintf("%x", hashBytes[:n])
}

func recordContainerInfo(cPID int, commandArray []string, cName string) (string, error) {
	// 首先生成容器ID
	id := randStringContainerID(IDLen)
	// 以当前时间为容器的创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	// 如果用户没有指定容器名就用容器ID做为容器名
	if cName == "" {
		cName = id
	}
	containerInfo := ContainerInfo{
		Pid:         strconv.Itoa(cPID),
		Id:          id,
		Name:        cName,
		Command:     strings.Join(commandArray, ""),
		CreatedTime: createTime,
		Status:      RUNNING,
	}
	// 序列为json
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.LogErrorFrom("recordContainerInfo", "Marshal", err)
		return "", err
	}
	// 创建容器信息对应的文件夹
	dirUrl := filepath.Join(DefaultInfoLocation, id)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.LogErrorFrom("recordContainerInfo", "MkdirAll", err)
		return "", err
	}
	// 创建json文件
	fileName := filepath.Join(dirUrl, ConfigName)
	configFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.LogErrorFrom("recordContainerInfo", "OpenFile", err)
		return "", err
	}
	defer configFile.Close()
	// 写入到文件
	if _, err := configFile.WriteString(string(jsonBytes)); err != nil {
		log.LogErrorFrom("recordContainerInfo", "WriteString", err)
		return "", err
	}
	return id, nil
}

func DeleteContainerInfo(containerID string) {
	dirUrl := filepath.Join(DefaultInfoLocation, containerID)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.LogErrorFrom("DeleteContainerInfo", "RemoveAll", err)
	}
}
