package arp

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
)

// Converts a MAC address to a byte slice
func MacToBytes(mac net.HardwareAddr) []byte {
	return []byte(mac)
}

// Converts an IP address to a byte slice (IPv4)
func IpToBytes(ip net.IP) []byte {
	return ip.To4()
}

// htons converts a uint16 from host to network byte order
func Htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

// BuildPacket constructs a complete Ethernet + ARP packet
func BuildPacket(senderMAC, targetMAC net.HardwareAddr, senderIP, targetIP net.IP, opcode uint16) []byte {
	packet := make([]byte, 42) // Ethernet header (14 bytes) + ARP packet (28 bytes)

	// Ethernet header
	copy(packet[0:6], targetMAC)
	copy(packet[6:12], senderMAC)
	packet[12] = 0x08
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

// Send sends a raw ARP packet using a raw socket
func Send(iface *net.Interface, packet []byte, targetMAC net.HardwareAddr) error {
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(Htons(0x0806)))
	if err != nil {
		return fmt.Errorf("socket error: %v", err)
	}
	defer syscall.Close(fd)

	addr := syscall.SockaddrLinklayer{
		Protocol: Htons(0x0806),
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
