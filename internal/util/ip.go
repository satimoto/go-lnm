package util

import (
	"errors"
	"net"
)

func GetIPAddress() (net.IP, error) {
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if addrs, err := iface.Addrs(); err == nil {
				for _, addr := range addrs {
					switch v := addr.(type) {
					case *net.IPNet:
						if !v.IP.IsLoopback() {
							return v.IP, nil
						}
					case *net.IPAddr:
						if !v.IP.IsLoopback() {
							return v.IP, nil
						}
					}
				}
			}
		}
	}

	return net.IP{}, errors.New("No IP address found")
}
