package subsystems

import "strings"

type ResourceConfig struct {
	MemoryLimit string // 内存限制
	CpuShare    string // CPU时间片权重
	CpuSet      string // CPU核心数
	CpuMems     string // CPU Node内存
}

// Subsystem 子系统统一接口，每个子系统都实现如下四个方法
// 这里cgroup抽象成为了path，因为cgroup在层级树的路径就是虚拟文件系统的路径
type Subsystem interface {
	Name() string                      // 返回子系统的名字
	Set(string, *ResourceConfig) error // 设置某个cgroup在这个子系统中的资源限制（设置子系统限制文件的内容）
	Apply(string, int) error           // 将进程添加到某个cgroup中
	Remove(string) error               // 移除某个cgroup
}

var (
	// SubsystemsIns 通过不同的子系统初始化实例创建资源限制的处理链数组
	SubsystemsIns = []Subsystem{
		&CpuSetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)

func (r *ResourceConfig) String() string {
	var line []string
	line = append(line, "MemoryLimit:", r.MemoryLimit)
	line = append(line, "CpuShare:", r.CpuShare)
	line = append(line, "CpuSet:", r.CpuSet)
	return strings.Join(line, " ")
}
