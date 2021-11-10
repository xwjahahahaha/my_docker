package subsystems

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"xwj/mydocker/log"
)

const (
	MemLimitFileName = "memory.limit_in_bytes"
	TaskFileName = "tasks"
)

var memoryLogger = log.Log.WithFields(logrus.Fields{
	"subsystem" : "memory",
})

type MemorySubSystem struct {

}

func (m *MemorySubSystem) Name() string {
	return "memory"
}

func (m *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	// GetCgroupPath获取当前子系统在虚拟文件系统中的路径
	subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		memoryLogger.WithFields(logrus.Fields{
			"method" : "Set",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	// 设置这个cgrouop的内存限制，将内存限制写入cgroup对应目录的memory.limit_in_bytes文件中
	if res.MemoryLimit != "" {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, MemLimitFileName), []byte(res.MemoryLimit), 0644); err != nil {
			memoryLogger.WithFields(logrus.Fields{
				"method" : "Set",
				"errFrom" : "WriteFile",
			}).Error(err)
			return err
		}
	}
	return nil
}

func (m *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false)
	if err != nil {
		memoryLogger.WithFields(logrus.Fields{
			"method" : "Apply",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	// 将进程的PID写入cgroup的虚拟文件系统对应的目录下的"task"文件夹
	if err := ioutil.WriteFile(path.Join(subsysCgroupPath, TaskFileName), []byte(strconv.Itoa(pid)), 0644); err != nil {
		memoryLogger.WithFields(logrus.Fields{
			"method" : "Apply",
			"errFrom" : "WriteFile",
		}).Error(err)
		return err
	}
	return nil
}


func (m *MemorySubSystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false)
	if err != nil {
		memoryLogger.WithFields(logrus.Fields{
			"method" : "Remove",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	// 删除掉cgroup的目录就是对整个cgroup的删除
	if err := os.RemoveAll(subsysCgroupPath); err != nil {
		memoryLogger.WithFields(logrus.Fields{
			"method" : "Remove",
			"errFrom" : "os.Remove",
		}).Error(err)
		return err
	}
	return nil
}

