package register

import (
	"go-cron/app/common"
	"net"
)

func getLocalIP() (string, error) {
	var (
		ipv4    string
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
		err     error
	)
	//获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return "", err
	}
	//获取第一个非lo的网卡
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return ipv4, nil
			}
		}
	}

	err = common.Err_No_Local_IP_Found
	return "", err
}
