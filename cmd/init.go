package cmd

func init() {
	rootCmd.AddCommand(initDocker, runDocker)
	runDocker.Flags().BoolVarP(&tty, "tty", "t", false, "enable tty")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.MemoryLimit, "memory-limit", "m", "200m", "memory limit")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.CpuShare, "cpu-shares", "", "1024", "cpu shares")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.CpuSet, "cpu-set", "", "0", "cpu set")
	runDocker.Flags().StringVarP(&ResourceLimitCfg.CpuMems, "cpu-mems", "", "0", "cpu memory")
	runDocker.Flags().StringVarP(&Volume, "volume", "v", "", "add a volume")
}