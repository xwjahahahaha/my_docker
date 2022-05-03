package network

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"xwj/mydocker/log"
)

const ipamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string            // 分配文件存放的位置
	Subnets             map[string]string // 网段和位图算法的数组map：key是网段，value是分配的位图字符串（使用字符串的一个字符标识一个状态位）
}

// 初始化一个IPAM对象
var ipAllocator = &IPAM{
	// 默认使用上面的默认存储位置作为分配信息存储位置
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

// load 加载IPAM的位图信息
func (ipam *IPAM) load() error {
	// 检查存储文件的状态，如果不存在则说明之前没有分配，则不需要加载
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	// 打开文件
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	// 读取文件
	contentBytes, err := ioutil.ReadAll(subnetConfigFile)
	// 反序列化
	err = json.Unmarshal(contentBytes, &ipam.Subnets)
	if err != nil {
		log.Log.Errorf("Error dump allocation info, %v", err)
		return err
	}
	return nil
}

// dump 存储IPAM的地址分配位图信息
func (ipam *IPAM) dump() error {
	// 检查存储文件所在文件夹是否存在,不存在则创建
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(ipamConfigFileDir, 0644); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// 打开文件
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	// json序列化
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	// 写入到文件
	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}
	return nil
}

// Allocate bitmap算法分配一个地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的map
	ipam.Subnets = make(map[string]string)
	// 从文件中加载已经分配的网段信息
	if err := ipam.load(); err != nil {
		log.Log.Error(err)
	}
	// 这里重新生成一个子网段实例，因为传递的是指针，为了避免影响
	_, subnet, _ = net.ParseCIDR(subnet.String())
	// subnet.Mask.Size() 函数会返回网段前面的固定位1的长度以及后面0位的长度
	// 例如: 127.0.0.0/8 子网掩码：255.0.0.0 subnet.Mask.Size()返回8和32
	one, size := subnet.Mask.Size()
	ipAddr := subnet.String()
	// 如果之前没有分配过这个网段，则初始化网段的分配配置
	if _, has := ipam.Subnets[ipAddr]; !has {
		// 1 << uint8(zero) 表示 2^(zero) 表示 剩余的可分配的IP数量，后面的位数全部用0填满
		ipam.Subnets[ipAddr] = strings.Repeat("0", 1<<uint8(size-one))
	}
	var AlloIP net.IP
	// 遍历网段的位图数组
	for c := range ipam.Subnets[ipAddr] {
		if ipam.Subnets[ipAddr][c] == '0' {
			// 设置这个为'0'的序号值为'1'，即表示分配这个IP
			// 转换字符串为字节数组进行修改
			ipalloc := []byte(ipam.Subnets[ipAddr])
			ipalloc[c] = '1'
			ipam.Subnets[ipAddr] = string(ipalloc)
			// 获取该网段的IP，比如对于网段192.168.0.0/16,这里就是192.168.0.0
			first_ip := subnet.IP
			// ip地址是一个uint[4]的数组，例如172.16.0.0就是[172, 16, 0, 0]
			// 需要通过数组中每一项加所需要的值, 对于当前序号,例如65535
			// 每一位加的计算就是[uint8(65535>>24), uint8(65535>>16), uint8(65535>>8), uint8(65535>>0)]
			for t := uint(4); t > 0; t -= 1 {
				[]byte(first_ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			// 由于这里从1开始分配，所以再加1
			first_ip[3] += 1
			AlloIP = first_ip
			break
		}
	}
	if err := ipam.dump(); err != nil {
		return nil, err
	}
	return AlloIP, nil
}

// Release 释放一个IP
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = make(map[string]string)
	_, subnet, _ = net.ParseCIDR(subnet.String())
	// 从文件中加载网段的分配信息
	if err := ipam.load(); err != nil {
		log.Log.Error(err)
		return err
	}
	// 计算IP地址再网段位图数组中的索引位置
	c := 0
	// 将IP地址转换为4个字节的表示方式
	releaseIP := ipaddr.To4()
	// 由于IP是从1开始分配的，所以转换成索引应减1
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		// 获取索引与分配IP相反：IP地址的每一位相减之后分别左移将对应的数值加到索引上
		// *8是IP的每一个小分段都是8位，等于扩大相应的倍数
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}
	// 将分配的位图数组中索引位置的值设置为0
	ipalloc := []byte(ipam.Subnets[subnet.String()])
	ipalloc[c] = '0'
	ipam.Subnets[subnet.String()] = string(ipalloc)
	// 保存释放掉IP之后的网段IP信息
	if err := ipam.dump(); err != nil {
		return err
	}
	return nil
}
