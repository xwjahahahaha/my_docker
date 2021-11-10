package cgroups

import (
	"github.com/sirupsen/logrus"
	"xwj/mydocker/cgroups/subsystems"
	"xwj/mydocker/log"
)

type CgroupManager struct {
	Path string								// cgroup在层级树中的路径，就是相对于系统层级树根cgroup目录的路径
	Resource *subsystems.ResourceConfig		// 资源配置
}

// NewCgroupManager
// @Description: 新建一个cgroup
// @param path
// @return *CgroupManager
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path:     path,
	}
}

// Apply
// @Description: 将当前进程放入各个子系统的cgroup中
// @receiver c
// @param pid
// @return error
func (c *CgroupManager) Apply(pid int) error {
	var errFlag bool
	for _, subSystemIns := range subsystems.SubsystemsIns {
		if err := subSystemIns.Apply(c.Path, pid); err != nil {
			log.Log.Errorf("process[%d] apply subsystem %s err.", pid, subSystemIns.Name())
			errFlag = true
		}
	}
	if !errFlag {
		log.Log.WithFields(logrus.Fields{
			"method" : "Apply",
		}).Infof("success apply process[%d] into cgroups", pid)
	}
	return nil
}

// Set
// @Description: 设置子系统限制
// @receiver c
// @param res
// @return error
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	var errFlag bool
	for _, subSystemIns := range subsystems.SubsystemsIns {
		if err := subSystemIns.Set(c.Path, res); err != nil {
			log.Log.Errorf("subsystem %s set limit err.", subSystemIns.Name())
			errFlag = true
		}
	}
	if !errFlag {
		log.Log.WithFields(logrus.Fields{
			"method" : "Set",
		}).Infof("success set limits:[%s] into those subsystems", res)
	}
	return nil
}

// Destroy
// @Description: 销毁各个子系统中的cgroup
// @receiver c
// @return error
func (c *CgroupManager) Destroy() error {
	var errFlag bool
	for _, subSystemIns := range subsystems.SubsystemsIns {
		if err := subSystemIns.Remove(c.Path); err != nil {
			log.Log.Errorf("subsystem %s remove cgroup err.", subSystemIns.Name())
			errFlag = true
		}
	}
	if !errFlag {
		log.Log.WithFields(logrus.Fields{
			"method" : "Destroy",
		}).Infof("success destroy cgroup %s files.", c.Path)
	}
	return nil
}