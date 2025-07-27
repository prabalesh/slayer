package shell

import (
	"fmt"
	"net"
	"strconv"

	"github.com/prabalesh/slayer/internal/networking"
	"github.com/prabalesh/slayer/internal/spoof"
)

// Modified Limit function for ShellSession
func (s *ShellSession) Limit(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: limit <host_id> <rate>")
		fmt.Println("  rate: e.g., '100kbit', '1mbit', '500kbit'")
		return
	}

	// Parse arguments
	hostId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid host ID")
		return
	}

	rate := args[1]

	// Get target host
	targetHost, exists := s.store.Hosts[int64(hostId)]
	if !exists {
		fmt.Println("Host not found")
		return
	}

	fmt.Printf("Limiting host %s (%s) to %s\n", targetHost.IP, targetHost.Hostname, rate)

	// Get gateway information
	gatewayIp, err := networking.GetDefaultGatewayIP()
	if err != nil {
		fmt.Println("Can't get the gateway IP")
		return
	}

	gatewayMAC, err := networking.GetGatewayMAC(gatewayIp)
	if err != nil {
		fmt.Printf("Can't get gateway MAC: %s\n", err)
		return
	}

	fmt.Printf("Gateway IP: %s, MAC: %s\n", gatewayIp, gatewayMAC)
	fmt.Printf("Interface: %s\n", s.store.Iface.Name)

	// Start ARP spoofing in a goroutine
	go spoof.Spoof(s.store.Iface, targetHost.IP, net.HardwareAddr(targetHost.MAC), gatewayIp, gatewayMAC)

}
