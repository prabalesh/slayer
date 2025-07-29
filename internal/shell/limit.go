package shell

import (
	"fmt"
	"strconv"
	"strings"
)

// Modified Limit function for ShellSession
func (s *ShellSession) Limit(args []string) {
	if len(args) < 3 {
		fmt.Println("❌ Usage: limit <host_id> <upload_rate|none> <download_rate|none>")
		fmt.Println("💡 Example: limit 1 100kbit 500kbit")
		fmt.Println("💡 Use 'none' if you want to skip upload/download limit")
		return
	}

	// Parse host ID
	hostId, err := strconv.Atoi(args[0])
	if err != nil || hostId < 0 {
		fmt.Printf("❌ Invalid host ID '%s': must be a positive number\n", args[0])
		return
	}

	uploadRate := strings.TrimSpace(args[1])
	downloadRate := strings.TrimSpace(args[2])

	// Validate rates
	if uploadRate == "none" {
		uploadRate = ""
	}
	if downloadRate == "none" {
		downloadRate = ""
	}
	if uploadRate == "" && downloadRate == "" {
		fmt.Println("❌ At least one of upload or download rate must be specified")
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
	fmt.Printf("⬆️  Upload Limit: %s\n", uploadRate)
	fmt.Printf("⬇️  Download Limit: %s\n", downloadRate)
	fmt.Printf("🔌 Interface: %s\n", s.store.Iface.Name)

	// Start ARP spoofing
	s.store.SpoofManager.Start(targetHost, s.store.Iface, s.store.GatewayIP, s.store.GatewayMAC)

	// Apply limit via limiter
	err = s.store.Limiter.Apply(targetHost.IP.String(), uploadRate, downloadRate)
	if err != nil {
		fmt.Printf("❌ Failed to apply rate limit: %v\n", err)
		return
	}

	s.store.Hosts[targetHost.ID].Limited = true
	s.store.Hosts[targetHost.ID].DownloadSpeed = downloadRate
	s.store.Hosts[targetHost.ID].UploadSpeed = uploadRate

	fmt.Printf("✅ Limit applied for %s (Up: %s, Down: %s)\n", targetHost.IP, uploadRate, downloadRate)
}
