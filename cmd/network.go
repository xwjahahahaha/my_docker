package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"xwj/mydocker/network"
)

var networkSubCMD = &cobra.Command{
	Use:   "network",
	Long:  "container network commands",
}

var networkCreateCMD = &cobra.Command{
	Use:   "create [network_name]",
	Short: "create a container network",
	Long:  "create a container network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载网络配置信息
		if err := network.Init(); err != nil {
			return err
		}
		// 创建网络
		if err := network.CreateNetwork(driver, subnet, args[0]); err != nil {
			return fmt.Errorf("create network error: %+v", err)
		}
		return nil
	},
}

var networkListCMD = &cobra.Command{
	Use:   "list",
	Short: "list container network",
	Long:  "list container network",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := network.Init(); err != nil {
			return err
		}
		network.ListNetwork()
		return nil
	},
}

var networkRemoveCMD = &cobra.Command{
	Use:   "remove [network_name]",
	Short: "remove container network",
	Long:  "remove container network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := network.Init(); err != nil {
			return err
		}
		if err := network.DeleteNetwork(args[0]); err != nil {
			return err
		}
		return nil
	},
}

