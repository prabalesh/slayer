package networking

import (
	"fmt"
	"net"
)

// getInterfaceCIDR returns the first IPv4 address in CIDR notation for the given interface
func GetInterfaceCIDR(iface *net.Interface) (string, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return ipnet.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv4 CIDR found on interface %s", iface.Name)
}

func GenerateIPsFromCIDR(cidr string) ([]net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []net.IP

	// Convert to IPv4 if applicable
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}

	// Start from network address
	ip = ip.Mask(ipnet.Mask)

	for ; ipnet.Contains(ip); inc(ip) {
		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)
		ips = append(ips, ipCopy)
	}

	// Remove network and broadcast addresses (only for IPv4 and if > 2 hosts)
	if ip.To4() != nil && len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

// inc increments an IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}
