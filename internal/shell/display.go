package shell

import "fmt"

func (s *ShellSession) DisplayActiveHosts() {
	if len(s.store.Hosts) <= 0 {
		fmt.Println("❌ No devices online")
		return
	}

	fmt.Println("\n📊 Active Hosts:")
	fmt.Println("══════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("%-4s %-15s %-18s %-30s %-8s\n", "ID", "IP Address", "MAC Address", "Hostname", "Limited")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────")

	for id, host := range s.store.Hosts {
		status := "❌"
		if host.Limited {
			status = "✅"
		}
		fmt.Printf("%-4d %-15s %-18s %-30s %-8s\n", id, host.IP, host.MAC, host.Hostname, status)
	}

	fmt.Println("══════════════════════════════════════════════════════════════════════════════")
	fmt.Printf("📈 Total devices found: %d\n\n", len(s.store.Hosts))
}
