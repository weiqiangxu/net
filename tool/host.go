package tool

import (
	"fmt"
	"net"
	"strconv"
)

// isValidIP check addr is ip
func isValidIP(addr string) bool {
	ip := net.ParseIP(addr)
	// 报告ip是否为全局单播
	// ip是否接口本地多播地址
	return ip.IsGlobalUnicast() && !ip.IsInterfaceLocalMulticast()
}

// getPort return a real port.
func getPort(lis net.Listener) (int, bool) {
	if addr, ok := lis.Addr().(*net.TCPAddr); ok {
		return addr.Port, true
	}
	return 0, false
}

// Extract returns a private addr and port.
func Extract(hostPort string, lis net.Listener) (string, error) {
	addr, port, err := net.SplitHostPort(hostPort)
	if err != nil && lis == nil {
		return "", err
	}
	if lis != nil {
		if p, ok := getPort(lis); ok {
			port = strconv.Itoa(p)
		} else {
			return "", fmt.Errorf("failed to extract port: %v", lis.Addr())
		}
	}
	if len(addr) > 0 && (addr != "0.0.0.0" && addr != "[::]" && addr != "::") {
		return net.JoinHostPort(addr, port), nil
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	lowest := int(^uint(0) >> 1)
	var result net.IP
	for _, i := range interfaces {
		if (i.Flags & net.FlagUp) == 0 {
			continue
		}
		if i.Index < lowest || result == nil {
			lowest = i.Index
		} else if result != nil {
			continue
		}
		adders, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, rawAddr := range adders {
			var ip net.IP
			switch addr := rawAddr.(type) {
			case *net.IPAddr:
				ip = addr.IP
			case *net.IPNet:
				ip = addr.IP
			default:
				continue
			}
			if isValidIP(ip.String()) {
				result = ip
			}
		}
	}
	if result != nil {
		return net.JoinHostPort(result.String(), port), nil
	}
	return "", nil
}
