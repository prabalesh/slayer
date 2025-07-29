package shell

import (
	"fmt"
	"strconv"
)

func (s *ShellSession) Unlimit(args []string) {
	if len(args) < 1 {
		fmt.Println("❌ Usage: unlimit <host_id>")
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

	// Check if host exists
	host, exists := s.store.Hosts[int64(hostId)]
	if !exists {
		fmt.Printf("❌ Host with ID %d not found\n", hostId)
		fmt.Println("💡 Use 'list' command to see available hosts")
		return
	}

	// Check if host is actually limited
	if !host.Limited {
		fmt.Printf("⚠️  Host %s (%s) is not currently limited\n", host.IP, host.Hostname)
		return
	}

	fmt.Printf("🔓 Removing bandwidth limit for %s (%s)...\n", host.IP, host.Hostname)

	// Remove bandwidth limit
	err = s.store.Limiter.Remove(host.IP.String(), s.store.Iface.Name)
	if err != nil {
		fmt.Printf("❌ Failed to remove bandwidth limit for %s: %v\n", host.IP, err)
		return
	}

	// Update host status
	s.store.Hosts[int64(hostId)].Limited = false
	s.store.Hosts[int64(hostId)].DownloadSpeed = ""
	s.store.Hosts[int64(hostId)].UploadSpeed = ""

	// Stop spoofing
	s.store.SpoofManager.Stop(int64(hostId))

	fmt.Printf("✅ Successfully unlimited host %s (%s)\n", host.IP, host.Hostname)
}
