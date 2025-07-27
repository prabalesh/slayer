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

// Mutex to prevent concurrent modifications
var mu sync.Mutex

// Track which IPs have download limits applied
var downloadLimitsApplied = make(map[string]bool)

// Constants
const (
	DownloadMark = "10"
	UploadMark   = "20"
	IFBDevice    = "ifb0"
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

// setupIFB0 initializes the ifb0 device for ingress shaping
func setupIFB0(iface string) error {
	// Load ifb module
	if err := runCommand("modprobe", "ifb"); err != nil {
		log.Printf("Warning: failed to load ifb module: %v", err)
	}

	// Create and bring up ifb0 (ignore errors if it already exists)
	runCommandIgnoreError("ip", "link", "add", IFBDevice, "type", "ifb")
	if err := runCommand("ip", "link", "set", "dev", IFBDevice, "up"); err != nil {
		return fmt.Errorf("failed to bring up %s: %v", IFBDevice, err)
	}

	// Setup ingress redirection from real interface to ifb0
	runCommandIgnoreError("tc", "qdisc", "del", "dev", iface, "ingress")
	if err := runCommand("tc", "qdisc", "add", "dev", iface, "handle", "ffff:", "ingress"); err != nil {
		return fmt.Errorf("failed to add ingress qdisc on %s: %v", iface, err)
	}

	if err := runCommand("tc", "filter", "add", "dev", iface, "parent", "ffff:", "protocol", "ip", "u32",
		"match", "u32", "0", "0", "action", "mirred", "egress", "redirect", "dev", IFBDevice); err != nil {
		return fmt.Errorf("failed to add ingress filter on %s: %v", iface, err)
	}

	// Add root qdisc to ifb0
	runCommandIgnoreError("tc", "qdisc", "del", "dev", IFBDevice, "root")
	if err := runCommand("tc", "qdisc", "add", "dev", IFBDevice, "root", "handle", "1:", "htb", "default", "30"); err != nil {
		return fmt.Errorf("failed to add root qdisc on %s: %v", IFBDevice, err)
	}

	return nil
}

// setupRootQdisc adds root qdisc to the real interface for upload shaping
func setupRootQdisc(iface string) error {
	runCommandIgnoreError("tc", "qdisc", "del", "dev", iface, "root")
	if err := runCommand("tc", "qdisc", "add", "dev", iface, "root", "handle", "1:", "htb", "default", "30"); err != nil {
		return fmt.Errorf("failed to add root qdisc on %s: %v", iface, err)
	}
	return nil
}

// Apply bandwidth limits to an IP address
func Apply(ip, uploadRate, downloadRate, iface string) error {
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
	if err := validateInterface(iface); err != nil {
		return err
	}

	// Generate class IDs
	downloadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "down"))
	uploadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "up"))

	// Setup ifb0 for download shaping if needed
	if downloadRate != "" {
		if err := setupIFB0(iface); err != nil {
			return err
		}
		downloadLimitsApplied[ip] = true
	}

	// Setup root qdisc for upload shaping
	if uploadRate != "" {
		if err := setupRootQdisc(iface); err != nil {
			return err
		}
	}

	// Set iptables mangle rules for upload only (download doesn't work with marks on ifb0)
	if uploadRate != "" {
		// Remove existing rule first (ignore errors)
		runCommandIgnoreError("iptables", "-t", "mangle", "-D", "PREROUTING", "-s", ip, "-j", "MARK", "--set-mark", UploadMark)
		if err := runCommand("iptables", "-t", "mangle", "-A", "PREROUTING", "-s", ip, "-j", "MARK", "--set-mark", UploadMark); err != nil {
			return fmt.Errorf("failed to add iptables upload rule for %s: %v", ip, err)
		}
	}

	// Apply DOWNLOAD limits (on ifb0) - use u32 filter instead of fw filter
	if downloadRate != "" {
		// Remove existing class and filter first (ignore errors)
		runCommandIgnoreError("tc", "filter", "del", "dev", IFBDevice, "protocol", "ip", "prio", "1", "u32", "match", "ip", "dst", ip)
		runCommandIgnoreError("tc", "class", "del", "dev", IFBDevice, "classid", downloadClass)

		if err := runCommand("tc", "class", "add", "dev", IFBDevice, "parent", "1:", "classid", downloadClass, "htb", "rate", downloadRate); err != nil {
			return fmt.Errorf("failed to add download class for %s: %v", ip, err)
		}

		// Use u32 filter to match destination IP directly instead of relying on iptables marks
		if err := runCommand("tc", "filter", "add", "dev", IFBDevice, "protocol", "ip", "prio", "1", "u32", "match", "ip", "dst", ip, "flowid", downloadClass); err != nil {
			return fmt.Errorf("failed to add download filter for %s: %v", ip, err)
		}
	}

	// Apply UPLOAD limits (on real interface)
	if uploadRate != "" {
		// Remove existing class and filter first (ignore errors)
		runCommandIgnoreError("tc", "filter", "del", "dev", iface, "protocol", "ip", "handle", UploadMark, "fw", "flowid", uploadClass)
		runCommandIgnoreError("tc", "class", "del", "dev", iface, "classid", uploadClass)

		if err := runCommand("tc", "class", "add", "dev", iface, "parent", "1:", "classid", uploadClass, "htb", "rate", uploadRate); err != nil {
			return fmt.Errorf("failed to add upload class for %s: %v", ip, err)
		}

		if err := runCommand("tc", "filter", "add", "dev", iface, "protocol", "ip", "handle", UploadMark, "fw", "flowid", uploadClass); err != nil {
			return fmt.Errorf("failed to add upload filter for %s: %v", ip, err)
		}
	}

	log.Printf("Successfully applied bandwidth limits for %s (upload: %s, download: %s)", ip, uploadRate, downloadRate)
	return nil
}

// Remove bandwidth limits from an IP address
func Remove(ip, iface string) error {
	mu.Lock()
	defer mu.Unlock()

	// Validate inputs
	if err := validateIP(ip); err != nil {
		return err
	}
	if err := validateInterface(iface); err != nil {
		return err
	}

	// Generate class IDs
	downloadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "down"))
	uploadClass := fmt.Sprintf("1:%d", ipToClassID(ip, "up"))

	// Remove iptables mangle rules (only upload uses marks)
	runCommandIgnoreError("iptables", "-t", "mangle", "-D", "PREROUTING", "-s", ip, "-j", "MARK", "--set-mark", UploadMark)

	// Remove tc download filter + class (from ifb0 if download limits were applied)
	if downloadLimitsApplied[ip] {
		runCommandIgnoreError("tc", "filter", "del", "dev", IFBDevice, "protocol", "ip", "prio", "1", "u32", "match", "ip", "dst", ip)
		runCommandIgnoreError("tc", "class", "del", "dev", IFBDevice, "classid", downloadClass)
		delete(downloadLimitsApplied, ip)
	}

	// Remove tc upload filter + class (from real interface)
	runCommandIgnoreError("tc", "filter", "del", "dev", iface, "protocol", "ip", "handle", UploadMark, "fw", "flowid", uploadClass)
	runCommandIgnoreError("tc", "class", "del", "dev", iface, "classid", uploadClass)

	log.Printf("Successfully removed bandwidth limits for %s", ip)
	return nil
}

// Update modifies existing bandwidth limits for an IP
func Update(ip, uploadRate, downloadRate, iface string) error {
	// Remove existing limits first, then apply new ones
	if err := Remove(ip, iface); err != nil {
		log.Printf("Warning: failed to remove existing limits for %s: %v", ip, err)
	}
	return Apply(ip, uploadRate, downloadRate, iface)
}

// ListLimits returns a list of IPs that currently have limits applied
func ListLimits() []string {
	mu.Lock()
	defer mu.Unlock()

	var ips []string
	for ip := range downloadLimitsApplied {
		ips = append(ips, ip)
	}
	return ips
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
	runCommandIgnoreError("tc", "qdisc", "del", "dev", IFBDevice, "root")

	// Bring down and remove ifb0
	runCommandIgnoreError("ip", "link", "set", "dev", IFBDevice, "down")
	runCommandIgnoreError("ip", "link", "del", IFBDevice)

	// Clear tracking maps
	downloadLimitsApplied = make(map[string]bool)

	log.Println("Cleanup completed")
	return nil
}
