package shell

import (
	"fmt"
	"strconv"
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

	// Print interface and gateway info (already available in store)
	fmt.Printf("Gateway IP: %s, MAC: %s\n", s.store.GatewayIP, s.store.GatewayMAC)
	fmt.Printf("Interface: %s\n", s.store.Iface.Name)

	// Start spoofing using SpoofManager
	s.store.SpoofManager.Start(targetHost, s.store.Iface, s.store.GatewayIP, s.store.GatewayMAC)

	// You could apply rate limiting here with `tc` if needed, e.g.:
	// limiter.Apply(targetHost.IP, rate)
}
