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
	CpuSetCpusLimitFileName = "cpuset.cpus"
	CpuSetMemsLimitFileName = "cpuset.mems"
)

var cpuSetLogger = log.Log.WithFields(logrus.Fields{
	"subsystem" : "cpuset",
})

type CpuSetSubSystem struct {

}

func (cs *CpuSetSubSystem) Name() string {
	return "cpuset"
}

func (cs *CpuSetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, true)
	if err != nil {
		cpuSetLogger.WithFields(logrus.Fields{
			"method" : "Set",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	// 注意：需要先设置cpu节点的内存再设置其他，否则会报错：no space left on device
	if res.CpuMems != "" {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, CpuSetMemsLimitFileName), []byte(res.CpuMems), 0644); err != nil {
			cpuSetLogger.WithFields(logrus.Fields{
				"method" : "Set",
				"errFrom" : "WriteFile",
			}).Error(err)
			return err
		}
	}
	if res.CpuSet != "" {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, CpuSetCpusLimitFileName), []byte(res.CpuSet), 0644); err != nil {
			cpuSetLogger.WithFields(logrus.Fields{
				"method" : "Set",
				"errFrom" : "WriteFile",
			}).Error(err)
			return err
		}
	}
	return nil
}

func (cs *CpuSetSubSystem) Apply(cgroupPath string, pid int) error {
	subsysCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, false)
	if err != nil {
		cpuSetLogger.WithFields(logrus.Fields{
			"method" : "Apply",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	// 将进程的PID写入cgroup的虚拟文件系统对应的目录下的"task"文件夹
	if err := ioutil.WriteFile(path.Join(subsysCgroupPath, TaskFileName), []byte(strconv.Itoa(pid)), 0644); err != nil {
		cpuSetLogger.WithFields(logrus.Fields{
			"method" : "Apply",
			"errFrom" : "WriteFile",
		}).Error(err)
		return err
	}
	return nil
}


func (cs *CpuSetSubSystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := GetCgroupPath(cs.Name(), cgroupPath, false)
	if err != nil {
		cpuSetLogger.WithFields(logrus.Fields{
			"method" : "Remove",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	// 删除掉cgroup的目录就是对整个cgroup的删除
	if err := os.RemoveAll(subsysCgroupPath); err != nil {
		cpuSetLogger.WithFields(logrus.Fields{
			"method" : "Remove",
			"errFrom" : "os.Remove",
		}).Error(err)
		return err
	}
	return nil
}