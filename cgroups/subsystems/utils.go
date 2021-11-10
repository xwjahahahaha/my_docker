package subsystems

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
	"xwj/mydocker/log"
)

// FindCgroupMountpoint
// @Description: 找到某个子系统的层级树中cgroup根节点所在的目录
// @param subsystem
// @return string
func FindCgroupMountpoint(subsystem string) string {
	// 根据虚拟文件系统/proc查询当前进程挂载信息
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()
	// 扫描目录
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")		// 按空格分割
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			// 如果选项中有当前子系统。则返回第五项(下标4)即系统创建的子系统路径
			if opt == subsystem{
				// 在一些系统中, /sys/fs/cgroup/cpu改为了/sys/fs/cgroup/cpu,cpuacct，所以做一个判断
				if fields[4] == "/sys/fs/cgroup/cpu,cpuacct" {
					return "/sys/fs/cgroup/cpu"
				}
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Log.WithFields(logrus.Fields{
			"method" : "FindCgroupMountpoint",
			"errFrom" : "WithFields",
		}).Error(err)
		return ""
	}
	return ""
}

// GetCgroupPath
// @Description: 获得当前子系统下的cgroup在系统层级树的绝对路径, 如果这个cgroup文件夹没有，可以设置自动创建
// @param subsystem
// @param cPath
// @param autoCreate
// @return string
// @return error
func GetCgroupPath(subsystem string, cPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountpoint(subsystem)
	absolutePath := path.Join(cgroupRoot, cPath)
	// 如果有这个cgroup绝对路径的文件目录 或者 没有这个目录但是设置了自动创建
	if _, err := os.Stat(absolutePath); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			// 创建文件夹
			if err := os.Mkdir(absolutePath, 0755); err != nil {
				return "", fmt.Errorf("error create cgroup dir %v", err)
			}
			return absolutePath, nil
		}
		return absolutePath, nil
	}else {
		// 如果os.Stat是其他错误或者不存在cgroup目录但是也没有设置自动创建，则返回错误
		return "", fmt.Errorf("cgroup path error %v", err)
	}
}
