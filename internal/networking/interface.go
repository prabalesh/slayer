package networking

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func GetActiveWiFiInterface() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		// Skip if not up or is loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Heuristic: likely a Wi-Fi interface
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "wlan") || strings.Contains(name, "wifi") || strings.HasPrefix(name, "wl") {
			return &iface, nil
		}
	}

	return nil, fmt.Errorf("no active Wi-Fi interface found")
}

func GetInterfaceByName(interfaceName string) *net.Interface {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		log.Fatalf("could not get interface %s: %v", interfaceName, err)
	}
	return iface
}
