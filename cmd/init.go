package cmd

func init() {
	rootCmd.AddCommand(initDocker, runDocker)
	runDocker.Flags().BoolVarP(&tty, "tty", "t", false, "enable tty")
}