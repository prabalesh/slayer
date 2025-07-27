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
	startedTime := time.Now()
	fmt.Println("⏳ Scanning network...")
	arpScanner := scanner.NewArpScanner(&s.store)
	arpScanner.Scan(ips)
	timeTaken := time.Since(startedTime)
	fmt.Println("\n✅ Scan completed!")
	s.DisplayActiveHosts()
	fmt.Printf("⏱️  Time taken: %v\n", timeTaken)
}
