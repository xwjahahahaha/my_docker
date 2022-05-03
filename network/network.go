package network

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
	"xwj/mydocker/log"
	"xwj/mydocker/record"
)

// Network 网络
type Network struct {
	Name    string     `json:"name"`     // 网络名
	IpRange *net.IPNet `json:"ip_range"` // 地址段
	Driver  string     `json:"driver"`   // 网络驱动名
}

// Endpoint 网络端点
type Endpoint struct {
	ID          string           `json:"id"`           // ID
	Device      netlink.Veth     `json:"dev"`          // Veth设备
	IpAddress   net.IP           `json:"ip"`           // IP地址
	MacAddress  net.HardwareAddr `json:"mac"`          // mac地址
	PortMapping []string         `json:"port_mapping"` // 端口映射
	Network     *Network         // 网络
}

// NetworkDriver 网络驱动
type NetworkDriver interface {
	Name() string                                          // 驱动名
	Create(subnet string, name string) (*Network, error)   // 创建网络
	Delete(network *Network) error                         // 删除网络
	Connect(network *Network, endpoint *Endpoint) error    // 连接容器网络端点到网络
	Disconnect(network *Network, endpoint *Endpoint) error // 从网络中移除容器的网络端点
}

var (
	defaultNetworkPath = "/var/run/mydocker/network/network/" // 默认存储位置
	drivers            = map[string]NetworkDriver{}           // 网络驱动映射
	networks           = map[string]*Network{}                // 所有网络映射
)

// CreateNetwork 根据网络驱动创建网络
func CreateNetwork(driver, subnet, name string) error {
	// ParseCIDR的功能是将网段的字符串转换为net.IPNet对象
	_, cidr, _ := net.ParseCIDR(subnet)
	// 通过IPAM分配网关IP，获取到网段中第一个IP作为网关的IP
	gatewayIp, err := ipAllocator.Allocate(cidr)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	// 重置IP
	cidr.IP = gatewayIp

	// 调用指定的网络驱动创建网络，这里的drivers字典是各个网络驱动的示例字典，通过调用网络驱动的Create方法创建网络
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	// 保存网络信息，将网络信息保存在文件系统中，以便查询和在网络上连接网络端点
	return nw.dump(defaultNetworkPath)
}

// Connect 容器连接网络
func Connect(networkName string, cinfo *record.ContainerInfo) error {
	// 从networks字典中获取容器连接的网络信息，networks字典中保存了当前已经创建的网络
	network, ok := networks[networkName]
	if !ok {
		err := fmt.Errorf(" No Such Network: %s", networkName)
		log.Log.Error(err)
		return err
	}
	// 通过调用IPAM从网络的网段中获取可用的IP作为容器IP地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IpAddress:   ip,
		PortMapping: cinfo.PortMapping,
		Network:     network,
	}
	// 调用网络驱动的Connect方法连接和配置网络端点
	if err := drivers[network.Driver].Connect(network, ep); err != nil {
		log.Log.Error(err)
		return err
	}
	// 进入到容器的网络Namespace配置容器网络设备的IP地址和路由
	if err := configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		log.Log.Error(err)
		return err
	}
	// 配置容器到宿主机的端口映射
	return configPortMapping(ep)
}

// Init 从网络配置的目录中加载所有的网络配置信息到networks字典中
func Init() error {
	// 加载网络驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	// 判断网络的配置目录是否存在，不存在则创建
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(defaultNetworkPath, 0644); err != nil {
				log.Log.Error(err)
				return err
			}
		} else {
			log.Log.Error(err)
			return err
		}
	}
	// 检查网络配置目录中的所有文件
	if err := filepath.Walk(defaultNetworkPath, func(nwPath string, info fs.FileInfo, err error) error {
		// 如果是目录则跳过
		if info.IsDir() {
			return nil
		}
		// 加载文件名作为网络名
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		// 调用Network.load方法加载网络配置信息
		if err := nw.load(nwPath); err != nil {
			log.Log.Error(err)
			return err
		}
		// 将网络配置信息加入到networks字典中
		networks[nwName] = nw
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// ListNetwork 遍历网络字典展示
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, v := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			v.Name,
			v.IpRange.String(),
			v.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		log.Log.Error(err)
		return
	}
}

func DeleteNetwork(networkName string) error {
	// 查找网络是否存在
	nw, ok := networks[networkName]
	if !ok {
		err := fmt.Errorf(" No Such Network: %s", networkName)
		log.Log.Error(err)
		return err
	}
	// 调用IPAM的实例释放网络网关的IP
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf(" Error Remove Network gateway ip: %s", err)
	}
	// 调用网络驱动删除网络创建的设备与配置
	if err := drivers[nw.Driver].Delete(nw); err != nil {
		return fmt.Errorf(" Error Remove Network DriverError: %s", err)
	}
	// 从网络的配置目录中删除该网络对应的配置文件
	return nw.remove(defaultNetworkPath)
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *record.ContainerInfo) error {
	// 获取网络端点中的Veth的另一端
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	// 将容器的网络端点加入到容器的网络空间中，并使这个函数下面的操作都在这个网络空间中进行，执行完函数后，恢复为默认的网络空间
	defer enterContainerNetns(&peerLink, cinfo)()
	// 获取到容器的IP地址以及网段，用于配置容器内部接口地址
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IpAddress
	// 调用setInterfaceIp函数设置容器内Veth端点的IP
	if err := setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("NetWork : %v, err : %s", ep.Network, err)
	}
	// 启动容器内的Veth端点
	if err := setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	// Net Namespace中默认本地地址127.0.0.1的lo网卡是关闭状态的，启动以保证容器访问自己的请求
	if err := setInterfaceUP("lo"); err != nil {
		return err
	}
	// 设置容器内的外部请求都通过容器内的Veth端点访问
	// 0.0.0.0/0的网段，表示所有的IP地址段
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	// 构建要添加的路由数据，包括网络设备，网关IP以及目的网段
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}
	return nil
}

// enterContainerNetns 进入容器内部并配置veth
// 锁定当前程序执行的线程，防止goroutine别调度到其他线程，离开目标网络空间
// 返回一个函数指针，执行这个返回函数才会退出容器的网络空间，回到宿主机的网络空间
func enterContainerNetns(enLink *netlink.Link, cinfo *record.ContainerInfo) func() {
	// 找到容器的Net Namespace
	// 通过/proc/[pid]/ns/net文件的文件描述符可以来操作Net Namepspace
	// pid就是containerInfo中的容器在宿主机上映射的进程ID
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Log.Error(err)
	}
	// 获取文件描述符
	nsFD := f.Fd()
	// 锁定当前线程
	runtime.LockOSThread()
	// 修改网络端点Veth的另一端，将其移动到容器的Net namespace中
	if err := netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.Log.Error(err)
	}
	// 通过netns.Get方法获得当前网络的Net Namespace，以便后面从容器的Net Namespace中退出回到当前netns中
	origins, err := netns.Get()
	if err != nil {
		log.Log.Error(err)
	}
	// 调用netns.Set方法，将当前进程加入容器的Net Namespace
	if err := netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Log.Error(err)
	}
	// 返回之前的Net Namespace函数
	// 在容器的网络空间中，执行完容器配置之后调用此函数就可以回到原来的Net Namespace
	return func() {
		// 恢复上面获取到的Net Namespace
		if err := netns.Set(origins); err != nil {
			log.Log.Error(err)
		}
		// 关闭Namespace文件
		origins.Close()
		// 取消线程锁定
		runtime.UnlockOSThread()
		// 关闭Namespace文件
		f.Close()
	}
}

// configPortMapping 配置端口映射
func configPortMapping(ep *Endpoint) error {
	// 遍历容器端口映射列表
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Log.Errorf("port mapping format error, %v", pm)
			continue
		}
		// 使用命令行实现iptable
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IpAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}

//dump 将网络配置信息保存在文件系统中
func (nw *Network) dump(dumpPath string) error {
	// 检查保存的目录是否存在
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dumpPath, 0644); err != nil {
				log.Log.Error(err)
				return err
			}
		} else {
			log.Log.Error(err)
			return err
		}
	}
	// 保存的文件名使用网络的名字
	nwPath := path.Join(dumpPath, nw.Name)
	// 打开保存的文件用于写入
	file, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	defer file.Close()

	// 通过json序列化
	nwBytes, err := json.Marshal(nw)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	// 写入
	if _, err := file.Write(nwBytes); err != nil {
		log.Log.Error(err)
		return err
	}
	return nil
}

// load 将配置文件信息加载到Network对象
func (nw *Network) load(dumpPath string) error {
	// 打开配置文件
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	defer nwConfigFile.Close()
	// 从配置文件中读取网络的配置json
	jsonBytes, err := ioutil.ReadAll(nwConfigFile)
	if err != nil {
		log.Log.Error(err)
		return err
	}
	// 反序列化
	if err := json.Unmarshal(jsonBytes, &nw); err != nil {
		log.Log.Error(err)
		return err
	}
	return nil
}

// remove 删除网络配置文件
func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}
