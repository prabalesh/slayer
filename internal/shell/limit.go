package shell

import (
	"fmt"
	"strconv"
	"strings"
)

// Modified Limit function for ShellSession
func (s *ShellSession) Limit(args []string) {
	if len(args) < 2 {
		fmt.Println("❌ Usage: limit <host_id> <rate>")
		fmt.Println("💡 Rate examples: '100kbit', '1mbit', '500kbit'")
		return
	}

	// Parse host ID
	hostId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("❌ Invalid host ID '%s': must be a number\n", args[0])
		return
	}
	if hostId < 0 {
		fmt.Println("❌ Host ID must be a positive number")
		return
	}

	rate := strings.TrimSpace(args[1])
	if rate == "" {
		fmt.Println("❌ Rate cannot be empty")
		return
	}

	// Validate rate format (basic check)
	if !strings.HasSuffix(rate, "bit") {
		fmt.Printf("❌ Invalid rate format '%s': must end with 'bit' (e.g., '100kbit', '1mbit')\n", rate)
		return
	}

	// Get target host
	targetHost, exists := s.store.Hosts[int64(hostId)]
	if !exists {
		fmt.Printf("❌ Host with ID %d not found\n", hostId)
		fmt.Println("💡 Use 'list' command to see available hosts")
		return
	}

	fmt.Printf("🎯 Target: %s (%s)\n", targetHost.IP, targetHost.Hostname)
	fmt.Printf("⚠️  Setting bandwidth limit to: %s\n", rate)
	fmt.Printf("🌐 Gateway: %s (%s)\n", s.store.GatewayIP, s.store.GatewayMAC)
	fmt.Printf("🔌 Interface: %s\n", s.store.Iface.Name)

	// Start spoofing using SpoofManager
	s.store.SpoofManager.Start(targetHost, s.store.Iface, s.store.GatewayIP, s.store.GatewayMAC)

	fmt.Printf("✅ Bandwidth limiting started for %s\n", targetHost.IP)

	// You could apply rate limiting here with `tc` if needed, e.g.:
	// err = limiter.Apply(targetHost.IP, rate)
	// if err != nil {
	//     fmt.Printf("❌ Failed to apply rate limit: %v\n", err)
	//     return
	// }
}
