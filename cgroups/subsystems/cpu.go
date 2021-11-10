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
	CpuShareLimitFileName = "cpu.shares"
)

var CpuSubLogger = log.Log.WithFields(logrus.Fields{
	"subsystem" : "cpu",
})

type CpuSubSystem struct {

}

func (c *CpuSubSystem) Name() string {
	return "cpu"
}

func (c *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true)
	if err != nil {
		CpuSubLogger.WithFields(logrus.Fields{
			"method" : "Set",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	if res.CpuShare != "" {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, CpuShareLimitFileName), []byte(res.CpuShare), 0644); err != nil {
			CpuSubLogger.WithFields(logrus.Fields{
				"method" : "Set",
				"errFrom" : "WriteFile",
			}).Error(err)
			return err
		}
	}
	return nil
}

func (c *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false)
	if err != nil {
		CpuSubLogger.WithFields(logrus.Fields{
			"method" : "Apply",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	if err := ioutil.WriteFile(path.Join(subsysCgroupPath, TaskFileName), []byte(strconv.Itoa(pid)), 0644); err != nil {
		CpuSubLogger.WithFields(logrus.Fields{
			"method" : "Apply",
			"errFrom" : "WriteFile",
		}).Error(err)
		return err
	}
	return nil
}


func (c *CpuSubSystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false)
	if err != nil {
		CpuSubLogger.WithFields(logrus.Fields{
			"method" : "Remove",
			"errFrom" : "GetCgroupPath",
		}).Error(err)
		return err
	}
	if err := os.RemoveAll(subsysCgroupPath); err != nil {
		CpuSubLogger.WithFields(logrus.Fields{
			"method" : "Remove",
			"errFrom" : "os.Remove",
		}).Error(err)
		return err
	}
	return nil
}