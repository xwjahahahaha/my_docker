package cmd

import "xwj/mydocker/cgroups/subsystems"

var (
	tty              bool                           // 是否交互式执行
	ResourceLimitCfg = &subsystems.ResourceConfig{} // 资源限制配置
	CgroupName       = "myDocker"                   // 新建的cgroup的名称
	Volume           string                         // 数据卷
	Detach           bool                           // 后台运行
	Name             string                         // 容器名称
	ImageTarPath     string                         // 镜像的tar包路径
	EnvSlice         []string                       // 环境变量
	NetWorkName      string                         // 网络名
	Port             []string                       // 端口映射

	driver string // 网络驱动名称
	subnet string // 子网网段
)

func init() {
	rootCMD.AddCommand(initContainerCMD, runContainerCMD, commitContainerCMD,
		listContainersCMD, logContainersCMD, execContainerCMD, stopContainerCMD,
		removeContainerCMD, networkSubCMD)
	networkSubCMD.AddCommand(networkCreateCMD, networkListCMD, networkRemoveCMD)

	runContainerCMD.Flags().BoolVarP(&tty, "tty", "t", false, "enable tty")
	runContainerCMD.Flags().StringVarP(&ResourceLimitCfg.MemoryLimit, "memory-limit", "m", "200m", "memory limit")
	runContainerCMD.Flags().StringVarP(&ResourceLimitCfg.CpuShare, "cpu-shares", "", "1024", "cpu shares")
	runContainerCMD.Flags().StringVarP(&ResourceLimitCfg.CpuSet, "cpu-set", "", "0", "cpu set")
	runContainerCMD.Flags().StringVarP(&ResourceLimitCfg.CpuMems, "cpu-mems", "", "0", "cpu memory")
	runContainerCMD.Flags().StringVarP(&Volume, "volume", "v", "", "add a volume")
	runContainerCMD.Flags().BoolVarP(&Detach, "detach", "d", false, "Run container in background and print container ID")
	runContainerCMD.Flags().StringVarP(&Name, "container-name", "n", "", "set a container nickname")
	runContainerCMD.Flags().StringVarP(&ImageTarPath, "image-tar-path", "i", "./busybox.tar", "used image tar file path")
	runContainerCMD.Flags().StringSliceVarP(&EnvSlice, "set-environment", "e", []string{}, "set environment")
	runContainerCMD.Flags().StringVarP(&NetWorkName, "net", "", "", "choose network")
	runContainerCMD.Flags().StringSliceVarP(&Port, "port-mapping", "p", []string{}, "set a port mapping")

	networkCreateCMD.Flags().StringVarP(&driver, "driver", "", "bridge", "network driver")
	networkCreateCMD.Flags().StringVarP(&subnet, "subnet", "", "", "subnet cidr")
	networkCreateCMD.MarkFlagRequired("driver")
	networkCreateCMD.MarkFlagRequired("subnet")
}
