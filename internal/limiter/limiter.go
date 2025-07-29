package limiter

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Limiter struct {
	iface *net.Interface
}

func NewLimiter() *Limiter {
	return &Limiter{}
}

func (l *Limiter) Init(iface *net.Interface) error {
	if err := runCommand("tc", "qdisc", "add", "dev", iface.Name, "root", "handle", "1:", "htb", "default", "999"); err != nil {
		return fmt.Errorf("failed to add root qdisc on %s: %v", iface.Name, err)
	}
	return nil
}

// Mutex to prevent concurrent modifications
var mu sync.Mutex

// Constants
const (
	DownloadMark = "10"
	UploadMark   = "20"
)

// validateIP checks if the IP address is valid
func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

// validateRate checks if the rate string is in valid tc format
func validateRate(rate string) error {
	if rate == "" {
		return nil // empty rate is allowed (means no limit)
	}
	// Basic validation for tc rate format (e.g., "1mbit", "100kbit", "1gbit")
	matched, _ := regexp.MatchString(`^\d+(bit|kbit|mbit|gbit|tbit|bps|kbps|mbps|gbps|tbps)$`, rate)
	if !matched {
		return fmt.Errorf("invalid rate format: %s (expected format like '1mbit', '100kbit')", rate)
	}
	return nil
}

// validateInterface checks if network interface exists
func validateInterface(iface string) error {
	_, err := net.InterfaceByName(iface)
	if err != nil {
		return fmt.Errorf("interface %s not found: %v", iface, err)
	}
	return nil
}

// runCommand executes a command and returns error if it fails
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command '%s %s' failed: %v", name, strings.Join(args, " "), err)
	}
	return nil
}

// runCommandIgnoreError executes a command but ignores errors (for cleanup operations)
func runCommandIgnoreError(name string, args ...string) {
	exec.Command(name, args...).Run()
}

// hash-based class ID based on IP
func ipToClassID(ip string, direction string) int {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 100 // fallback
	}
	lastOctet := parts[3]
	base := 100
	if direction == "up" {
		base = 200
	}
	id, err := strconv.Atoi(lastOctet)
	if err != nil {
		return base
	}
	return base + id
}

// Apply bandwidth limits to an IP address
func (l *Limiter) Apply(ip, uploadRate, downloadRate string) error {
	mu.Lock()
	defer mu.Unlock()

	// Validate inputs
	if err := validateIP(ip); err != nil {
		return err
	}
	if err := validateRate(uploadRate); err != nil {
		return err
	}
	if err := validateRate(downloadRate); err != nil {
		return err
	}

	// Generate class IDs
	downloadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "down"))
	uploadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "up"))

	// Set iptables mangle rules for upload only (download doesn't work with marks on ifb0)
	if uploadRate != "" {
		// Remove existing rule first (ignore errors)
		runCommandIgnoreError("iptables", "-t", "mangle", "-D", "PREROUTING", "-s", ip, "-j", "MARK", "--set-mark", UploadMark)
		if err := runCommand("iptables", "-t", "mangle", "-A", "PREROUTING", "-s", ip, "-j", "MARK", "--set-mark", UploadMark); err != nil {
			return fmt.Errorf("failed to add iptables upload rule for %s: %v", ip, err)
		}
	}

	if downloadRate != "" {
		// Remove existing rule first (ignore errors)
		runCommandIgnoreError("iptables", "-t", "mangle", "-D", "PREROUTING", "-s", ip, "-j", "MARK", "--set-mark", DownloadMark)
		if err := runCommand("iptables", "-t", "mangle", "-A", "PREROUTING", "-d", ip, "-j", "MARK", "--set-mark", DownloadMark); err != nil {
			return fmt.Errorf("failed to add iptables download rule for %s: %v", ip, err)
		}
	}

	if downloadRate != "" {
		// Remove existing class and filter first (ignore errors)
		// runCommandIgnoreError("tc", "filter", "del", "dev", iface, "protocol", "ip", "handle", DownloadMark, "fw", "flowid", downloadClass)
		// runCommandIgnoreError("tc", "class", "del", "dev", iface, "classid", downloadClass)

		if err := runCommand("tc", "class", "add", "dev", l.iface.Name, "parent", "1:", "classid", downloadClass, "htb", "rate", downloadRate); err != nil {
			if err := runCommand("tc", "class", "change", "dev", l.iface.Name, "parent", "1:", "classid", downloadClass, "htb", "rate", downloadRate); err != nil {
				return fmt.Errorf("failed to add download class for %s: %v", ip, err)
			}
		}

		if err := runCommand("tc", "filter", "add", "dev", l.iface.Name, "protocol", "ip", "handle", DownloadMark, "fw", "flowid", downloadClass); err != nil {
			return fmt.Errorf("failed to add upload filter for %s: %v", ip, err)
		}
	}

	// Apply UPLOAD limits (on real interface)
	if uploadRate != "" {
		// Remove existing class and filter first (ignore errors)
		// runCommandIgnoreError("tc", "filter", "del", "dev", l.iface.Name, "protocol", "ip", "handle", UploadMark, "fw", "flowid", uploadClass)
		// runCommandIgnoreError("tc", "class", "del", "dev", l.iface.Name, "classid", uploadClass)

		if err := runCommand("tc", "class", "add", "dev", l.iface.Name, "parent", "1:", "classid", uploadClass, "htb", "rate", uploadRate); err != nil {
			if err := runCommand("tc", "class", "change", "dev", l.iface.Name, "parent", "1:", "classid", uploadClass, "htb", "rate", uploadRate); err != nil {
				return fmt.Errorf("failed to add upload class for %s: %v", ip, err)
			}
		}

		if err := runCommand("tc", "filter", "add", "dev", l.iface.Name, "protocol", "ip", "handle", UploadMark, "fw", "flowid", uploadClass); err != nil {
			return fmt.Errorf("failed to add upload filter for %s: %v", ip, err)
		}
	}

	log.Printf("Successfully applied bandwidth limits for %s (upload: %s, download: %s)", ip, uploadRate, downloadRate)
	return nil
}

// Remove bandwidth limits from an IP address
func (l *Limiter) Remove(ip string) error {
	mu.Lock()
	defer mu.Unlock()

	// Validate inputs
	if err := validateIP(ip); err != nil {
		return err
	}
	if err := validateInterface(l.iface.Name); err != nil {
		return err
	}

	// Generate class IDs
	downloadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "down"))
	uploadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "up"))

	// Remove iptables mangle rules (only upload uses marks)
	runCommandIgnoreError("iptables", "-t", "mangle", "-D", "PREROUTING", "-s", ip, "-j", "MARK", "--set-mark", UploadMark)
	runCommandIgnoreError("iptables", "-t", "mangle", "-D", "PREROUTING", "-d", ip, "-j", "MARK", "--set-mark", DownloadMark)

	// Remove tc download filter + class (from ifb0 if download limits were applied)
	runCommandIgnoreError("tc", "filter", "del", "dev", l.iface.Name, "protocol", "ip", "handle", DownloadMark, "fw", "flowid", downloadClass)
	runCommandIgnoreError("tc", "class", "del", "dev", l.iface.Name, "classid", downloadClass)

	// Remove tc upload filter + class (from real interface)
	runCommandIgnoreError("tc", "filter", "del", "dev", l.iface.Name, "protocol", "ip", "handle", UploadMark, "fw", "flowid", uploadClass)
	runCommandIgnoreError("tc", "class", "del", "dev", l.iface.Name, "classid", uploadClass)

	log.Printf("Successfully removed bandwidth limits for %s", ip)
	return nil
}

// Cleanup removes all bandwidth limiting rules and cleans up interfaces
func Cleanup(iface string) error {
	mu.Lock()
	defer mu.Unlock()

	log.Println("Cleaning up all bandwidth limiting rules...")

	// Remove all iptables mangle rules
	runCommandIgnoreError("iptables", "-t", "mangle", "-F", "PREROUTING")
	runCommandIgnoreError("iptables", "-t", "mangle", "-F", "POSTROUTING")

	// Remove tc qdiscs
	runCommandIgnoreError("tc", "qdisc", "del", "dev", iface, "root")
	runCommandIgnoreError("tc", "qdisc", "del", "dev", iface, "ingress")

	log.Println("Cleanup completed")
	return nil
}
