package shell

import (
	"fmt"
	"log"
	"time"

	"github.com/prabalesh/slayer/internal/networking"
	"github.com/prabalesh/slayer/internal/scanner"
)

func (s *ShellSession) RunNetworkScan() {
	fmt.Printf("🌐 Detected interface: %s\n", s.store.Iface.Name)
	fmt.Printf("📍 Detected CIDR: %s\n", s.store.CIDR)

	ips, err := networking.GenerateIPsFromCIDR(s.store.CIDR)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("🎯 Scanning %d IPs using optimized ARP method (most accurate)...\n", len(ips))

	startedTime := time.Now()

	// Use the optimized ARP scanner for maximum accuracy
	optimizedScanner := scanner.NewArpScanner(&s.store)
	optimizedScanner.Scan(ips)

	timeTaken := time.Since(startedTime)

	fmt.Println("\n✅ Scan completed!")
	s.DisplayActiveHosts()
	fmt.Printf("⏱️  Time taken: %v (optimized for accuracy + performance)\n", timeTaken)
}
