package cmd

func init() {
	rootCmd.AddCommand(initDocker, runDocker, commitCommand, listContainers, logCommand, execCommand, stopCommand, removeCommand, networkCommand)
	networkCommand.AddCommand(networkCreate, networkList, networkRemove)

	runDocker.Flags().BoolVarP(&tty, "tty", "t", false, "enable tty")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.MemoryLimit, "memory-limit", "m", "200m", "memory limit")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.CpuShare, "cpu-shares", "", "1024", "cpu shares")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.CpuSet, "cpu-set", "", "0", "cpu set")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.CpuMems, "cpu-mems", "", "0", "cpu memory")
	runDocker.Flags().StringVarP(&Volume, "volume", "v", "", "add a volume")
	runDocker.Flags().BoolVarP(&Detach, "detach", "d", false, "Run container in background and print container ID")
	runDocker.Flags().StringVarP(&Name, "container-name", "n", "", "set a container nickname")
	runDocker.Flags().StringVarP(&ImageTarPath, "image-tar-path", "i", "./busybox.tar", "used image tar file path")
	runDocker.Flags().StringSliceVarP(&EnvSlice, "set-environment", "e", []string{}, "set environment")
	runDocker.Flags().StringVarP(&NetWorkName, "net", "", "", "choose network")
	runDocker.Flags().StringSliceVarP(&Port, "port-mapping", "p", []string{}, "set a port mapping")

	networkCreate.Flags().StringVarP(&driver, "driver", "", "bridge", "network driver")
	networkCreate.Flags().StringVarP(&subnet, "subnet", "", "", "subnet cidr")
	networkCreate.MarkFlagRequired("driver")
	networkCreate.MarkFlagRequired("subnet")
}