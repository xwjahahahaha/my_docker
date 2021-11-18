package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
	"xwj/mydocker/log"
)

// BridgeNetworkDriver Bridge网络驱动
type BridgeNetworkDriver struct {
}

// initBridge 初始化一个网桥
func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	// 1. 创建Bridge虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf(" Error add bridge： %s, Error: %v", bridgeName, err)
	}
	// 2. 设置Bridge设备的地址和路由
	gatewayIp := *n.IpRange
	gatewayIp.IP = n.IpRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIp.String()); err != nil {
		return fmt.Errorf(" Error assigning address: %s on bridge: %s with an error of: %v", gatewayIp, bridgeName, err)
	}
	// 3. 启动Bridge设备
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf(" Error set bridge up: %s, Error: %v", bridgeName, err)
	}

	// 4. 设置iptables的SNAT规则
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return fmt.Errorf(" Error setting iptables for %s: %v", bridgeName, err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	// 获取网段字符串的网关IP地址和网络IP段
	ip, ipRange, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Log.Error(err)
		return nil, err
	}
	ipRange.IP = ip
	// 初始化网络对象
	n := &Network{
		Name: name,
		IpRange: ipRange,
		Driver: d.Name(),
	}
	// 初始化配置Linux Bridge
	if err := d.initBridge(n); err != nil {
		log.Log.Error(err)
		return nil, err
	}
	// 返回配置好的网络
	return n, nil
}


// Delete 删除Bridge网络设备
func (d *BridgeNetworkDriver) Delete(network *Network) error {
	// 网络名即Linux Bridge的设备名
	bridgeName := network.Name
	// 通过netlink找到对应的设备
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf(" error get interface: %v", err)
	}
	// 删除网络对应的Linux Bridge设备
	return netlink.LinkDel(iface)
}

// Connect 创建Veth并连接网络与Veth网络端点
func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	// 网络名即Linux Bridge的设备名
	bridgeName := network.Name
	// 通过netlink找到对应的设备
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf(" error get interface: %v", err)
	}

	// 创建Veth接口的配置
	la := netlink.NewLinkAttrs()
	// 由于Linux接口名的限制，所以名字取前5位
	la.Name = endpoint.ID[:5]
	// 通过设置Veth接口的master属性，设置这个Veth的一端挂载到网络对应的Linux Bridge上
	la.MasterIndex = br.Attrs().Index

	// 创建Veth对象，通过PeerName配置Veth另一端的接口名cif-{endpoint ID前5位}
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName: "cif-" + endpoint.ID[:5],
	}

	// 调用netlink的LinkAdd方法创建出这个Veth接口
	// 因为上面已经指定了link的MasterIndex是网络接口Bridge，所以一端即已经挂载在Bridge上了
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf(" Error Add Endpoint Device: %v", err)
	}
	// 调用netlink的LinkSetUp方法设置Veth启动
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf(" Error set Endpoint Device up: %v", err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	// TODO 取消连接
	return nil
}

// createBridgeInterface 创建一个Bridge网络驱动/虚拟设备
func createBridgeInterface(bridgeName string) error {
	// 先检查是否存在同名的Bridge设备
	iface, err := net.InterfaceByName(bridgeName)
	if iface != nil || err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return fmt.Errorf(" Bridge interface %s exist!", bridgeName)
	}
	// 初始化一个netlink的link基础对象，link的名字就是bridge的名字
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	// 使用刚才创建的Link的属性创建netlink的Bridge对象
	br := &netlink.Bridge{ LinkAttrs: la }
	// 调用netlink的LinkAdd方法，创建Bridge虚拟网络设备
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf(" Bridge creation failed for bridge %s: %v", bridgeName, err)
	}
	return nil
}

// setInterfaceIP 设置Bridge设备的地址和路由
func setInterfaceIP(name string, rawIP string) error {
	// 通过netlink的LinkByName方法找到需要设置的网络接口也就是刚刚创建的Bridge
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf(" error get interface: %v", err)
	}
	// netlink.ParseIPNet是对net.ParseCIDR的封装，返回的值ipNet中既包含了网段的信息(192.168.0.0/24)也包含了原始的IP地址(192.168.0.1)
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	// 通过netlink.AddrAdd给网络接口配置地址，等价于 ip addr add xxxx命令
	// 同时如果配置了地址所在的网段信息，例如192.168.0.0/24, 还会配置路由表192.168.0.0/24转发到这个bridge上
	addr := &netlink.Addr{
		IPNet:       ipNet,
		Label:       "",
		Flags:       0,
		Scope:       0,
	}
	return netlink.AddrAdd(iface, addr)
}

// setInterfaceUP 设置网络接口(bridge)启动为Up状态
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf(" error get interface: %v", err)
	}
	// 通过netlink的LinkSetUp方法设置接口状态为Up状态
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf(" Error enabing interface for %s: %v", interfaceName, err)
	}
	return nil
}

// 设置iptable对应bridge的MASQUERADE规则
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// 由于go没有直接操作iptables的库，所以直接使用命令
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		log.Log.Errorf("iptables Output, %v", output)
	}
	return nil
}


