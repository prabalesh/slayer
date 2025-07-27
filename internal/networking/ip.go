package networking

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
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

// GetDefaultGatewayIP returns the default gateway as net.IP
func GetDefaultGatewayIP() (net.IP, error) {
	out, err := exec.Command("ip", "route").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ip route: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "default") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "via" && i+1 < len(fields) {
					ip := net.ParseIP(fields[i+1])
					if ip == nil {
						return nil, fmt.Errorf("failed to parse gateway IP: %s", fields[i+1])
					}
					return ip, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("default gateway not found")
}

// GetMACFromIP sends an ARP request to get the MAC address of the given net.IP
func GetGatewayMAC(gatewayIP net.IP) (net.HardwareAddr, error) {
	out, err := exec.Command("ip", "neigh").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ip neigh: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, gatewayIP.String()) {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "lladdr" && i+1 < len(fields) {
					macStr := fields[i+1]
					mac, err := net.ParseMAC(macStr)
					if err != nil {
						return nil, fmt.Errorf("invalid MAC address: %s", macStr)
					}
					return mac, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("MAC address not found for gateway IP: %s", gatewayIP.String())
}
