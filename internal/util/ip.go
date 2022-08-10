package util

import (
	"errors"
	"fmt"
	"net"
)

func GetIPAddress() (string, error) {
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if addrs, err := iface.Addrs(); err == nil {
				for _, addr := range addrs {
					switch v := addr.(type) {
					case *net.IPNet:
						if !v.IP.IsLoopback() {
							return formatIPAddress(v.IP), nil
						}
					case *net.IPAddr:
						if !v.IP.IsLoopback() {
							return formatIPAddress(v.IP), nil
						}
					}
				}
			}
		}
	}

	return "", errors.New("No IP address found")
}

func formatIPAddress(ip net.IP) string {
	if ip.To4() == nil {
		return fmt.Sprintf("[%s]", ip.String())
	}

	return ip.String()
}