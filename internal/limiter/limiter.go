package limiter

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// Direction constants
const (
	DirectionNone     = 0
	DirectionOutgoing = 1
	DirectionIncoming = 2
	DirectionBoth     = 3
)

// HostLimitIDs holds the TC class IDs for upload and download
type HostLimitIDs struct {
	UploadID   int
	DownloadID int
}

// LimitInfo stores information about a limited host
type LimitInfo struct {
	IDs       *HostLimitIDs
	Rate      string // e.g., "100kbit"
	Direction int
}

// TrafficLimiter manages bandwidth limiting for hosts
type TrafficLimiter struct {
	Interface string
	hostDict  map[int64]*LimitInfo // hostID -> LimitInfo
	mutex     sync.RWMutex
	nextID    int
}

// NewTrafficLimiter creates a new traffic limiter instance
func NewTrafficLimiter(iface string) *TrafficLimiter {
	limiter := &TrafficLimiter{
		Interface: iface,
		hostDict:  make(map[int64]*LimitInfo),
		nextID:    1,
	}

	// Initialize root qdisc
	limiter.initializeQdisc()

	return limiter
}

// initializeQdisc sets up the root HTB qdisc
func (tl *TrafficLimiter) initializeQdisc() {
	// Disable TCP segmentation offload (critical for traffic shaping)
	exec.Command("ethtool", "-K", tl.Interface, "tso", "off").Run()
	exec.Command("ethtool", "-K", tl.Interface, "gso", "off").Run()

	// Remove existing qdisc if present
	exec.Command("tc", "qdisc", "del", "dev", tl.Interface, "root").Run()

	// Add root HTB qdisc with default class
	cmd := exec.Command("tc", "qdisc", "add", "dev", tl.Interface, "root", "handle", "1:", "htb", "default", "30")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to initialize qdisc: %v\n", err)
		return
	}

	// Add root class (total bandwidth available)
	cmd = exec.Command("tc", "class", "add", "dev", tl.Interface, "parent", "1:", "classid", "1:1", "htb", "rate", "1000mbit")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to add root class: %v\n", err)
	}

	// Add default class for unclassified traffic
	cmd = exec.Command("tc", "class", "add", "dev", tl.Interface, "parent", "1:1", "classid", "1:30", "htb", "rate", "1mbit")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to add default class: %v\n", err)
	}
}

// generateIDs creates unique IDs for a host
func (tl *TrafficLimiter) generateIDs() *HostLimitIDs {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	uploadID := tl.nextID
	tl.nextID++
	downloadID := tl.nextID
	tl.nextID++

	return &HostLimitIDs{
		UploadID:   uploadID,
		DownloadID: downloadID,
	}
}

// calculateBurst computes appropriate burst size for a given rate
func calculateBurst(rate string) string {
	// Simple heuristic: add 'b' suffix and increase by ~10%
	if strings.HasSuffix(rate, "kbit") {
		if num, err := strconv.Atoi(strings.TrimSuffix(rate, "kbit")); err == nil {
			return fmt.Sprintf("%dkb", num/8+num/80) // Convert to bytes and add 10%
		}
	} else if strings.HasSuffix(rate, "mbit") {
		if num, err := strconv.Atoi(strings.TrimSuffix(rate, "mbit")); err == nil {
			return fmt.Sprintf("%dkb", num*125+num*12) // Convert to KB and add 10%
		}
	}
	return "15kb" // Default burst
}

// executeCommand runs a system command with better error handling
func executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command failed: %s %v\nOutput: %s\n", name, args, string(output))
	}
	return err
}

// LimitHost applies bandwidth limiting to a host
func (tl *TrafficLimiter) LimitHost(hostID int64, hostIP string, direction int, rate string) error {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	// Remove existing limits if any
	if info, exists := tl.hostDict[hostID]; exists {
		tl.unlimitHostUnsafe(hostID, hostIP, info)
	}

	// Generate new IDs
	ids := tl.generateIDs()

	// Apply outgoing (upload) limits
	if (direction & DirectionOutgoing) == DirectionOutgoing {
		// Calculate burst rate (typically 10% more than rate)
		burstRate := calculateBurst(rate)

		err := executeCommand("tc", "class", "add", "dev", tl.Interface,
			"parent", "1:1", "classid", fmt.Sprintf("1:%d", ids.UploadID),
			"htb", "rate", rate, "ceil", rate, "burst", burstRate)
		if err != nil {
			return fmt.Errorf("failed to add upload class: %v", err)
		}

		// Add filter for marked packets
		err = executeCommand("tc", "filter", "add", "dev", tl.Interface,
			"parent", "1:", "protocol", "ip", "prio", strconv.Itoa(ids.UploadID),
			"handle", strconv.Itoa(ids.UploadID), "fw", "flowid", fmt.Sprintf("1:%d", ids.UploadID))
		if err != nil {
			return fmt.Errorf("failed to add upload filter: %v", err)
		}

		// Mark outgoing packets - use FORWARD chain for bridged traffic
		err = executeCommand("iptables", "-t", "mangle", "-A", "FORWARD",
			"-s", hostIP, "-j", "MARK", "--set-mark", strconv.Itoa(ids.UploadID))
		if err != nil {
			// Fallback to POSTROUTING if FORWARD fails
			executeCommand("iptables", "-t", "mangle", "-A", "POSTROUTING",
				"-s", hostIP, "-j", "MARK", "--set-mark", strconv.Itoa(ids.UploadID))
		}
	}

	// Apply incoming (download) limits
	if (direction & DirectionIncoming) == DirectionIncoming {
		// Calculate burst rate
		burstRate := calculateBurst(rate)

		err := executeCommand("tc", "class", "add", "dev", tl.Interface,
			"parent", "1:1", "classid", fmt.Sprintf("1:%d", ids.DownloadID),
			"htb", "rate", rate, "ceil", rate, "burst", burstRate)
		if err != nil {
			return fmt.Errorf("failed to add download class: %v", err)
		}

		// Add filter
		err = executeCommand("tc", "filter", "add", "dev", tl.Interface,
			"parent", "1:", "protocol", "ip", "prio", strconv.Itoa(ids.DownloadID),
			"handle", strconv.Itoa(ids.DownloadID), "fw", "flowid", fmt.Sprintf("1:%d", ids.DownloadID))
		if err != nil {
			return fmt.Errorf("failed to add download filter: %v", err)
		}

		// Mark incoming packets - use FORWARD chain for bridged traffic
		err = executeCommand("iptables", "-t", "mangle", "-A", "FORWARD",
			"-d", hostIP, "-j", "MARK", "--set-mark", strconv.Itoa(ids.DownloadID))
		if err != nil {
			// Fallback to PREROUTING if FORWARD fails
			executeCommand("iptables", "-t", "mangle", "-A", "PREROUTING",
				"-d", hostIP, "-j", "MARK", "--set-mark", strconv.Itoa(ids.DownloadID))
		}
	}

	// Store limit info
	tl.hostDict[hostID] = &LimitInfo{
		IDs:       ids,
		Rate:      rate,
		Direction: direction,
	}

	return nil
}

// BlockHost blocks all traffic for a host
func (tl *TrafficLimiter) BlockHost(hostID int64, hostIP string, direction int) error {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	// Remove existing limits if any
	if info, exists := tl.hostDict[hostID]; exists {
		tl.unlimitHostUnsafe(hostID, hostIP, info)
	}

	// Generate IDs for tracking
	ids := tl.generateIDs()

	// Block outgoing traffic
	if (direction & DirectionOutgoing) == DirectionOutgoing {
		err := executeCommand("iptables", "-t", "filter", "-A", "FORWARD",
			"-s", hostIP, "-j", "DROP")
		if err != nil {
			return fmt.Errorf("failed to block outgoing traffic: %v", err)
		}
	}

	// Block incoming traffic
	if (direction & DirectionIncoming) == DirectionIncoming {
		err := executeCommand("iptables", "-t", "filter", "-A", "FORWARD",
			"-d", hostIP, "-j", "DROP")
		if err != nil {
			return fmt.Errorf("failed to block incoming traffic: %v", err)
		}
	}

	// Store block info (rate = "" indicates blocking)
	tl.hostDict[hostID] = &LimitInfo{
		IDs:       ids,
		Rate:      "", // Empty rate indicates blocking
		Direction: direction,
	}

	return nil
}

// UnlimitHost removes all limits/blocks from a host
func (tl *TrafficLimiter) UnlimitHost(hostID int64, hostIP string) error {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	info, exists := tl.hostDict[hostID]
	if !exists {
		return nil // Already unlimited
	}

	return tl.unlimitHostUnsafe(hostID, hostIP, info)
}

// unlimitHostUnsafe removes limits without locking (internal use)
func (tl *TrafficLimiter) unlimitHostUnsafe(hostID int64, hostIP string, info *LimitInfo) error {
	direction := info.Direction
	ids := info.IDs

	// Remove outgoing limits/blocks
	if (direction & DirectionOutgoing) == DirectionOutgoing {
		if info.Rate != "" { // Was limited, not blocked
			// Remove TC class and filter
			executeCommand("tc", "filter", "del", "dev", tl.Interface,
				"parent", "1:0", "prio", strconv.Itoa(ids.UploadID))
			executeCommand("tc", "class", "del", "dev", tl.Interface,
				"parent", "1:0", "classid", fmt.Sprintf("1:%d", ids.UploadID))
			// Remove iptables MARK rule
			executeCommand("iptables", "-t", "mangle", "-D", "POSTROUTING",
				"-s", hostIP, "-j", "MARK", "--set-mark", strconv.Itoa(ids.UploadID))
		} else { // Was blocked
			// Remove iptables DROP rule
			executeCommand("iptables", "-t", "filter", "-D", "FORWARD",
				"-s", hostIP, "-j", "DROP")
		}
	}

	// Remove incoming limits/blocks
	if (direction & DirectionIncoming) == DirectionIncoming {
		if info.Rate != "" { // Was limited, not blocked
			// Remove TC class and filter
			executeCommand("tc", "filter", "del", "dev", tl.Interface,
				"parent", "1:0", "prio", strconv.Itoa(ids.DownloadID))
			executeCommand("tc", "class", "del", "dev", tl.Interface,
				"parent", "1:0", "classid", fmt.Sprintf("1:%d", ids.DownloadID))
			// Remove iptables MARK rule
			executeCommand("iptables", "-t", "mangle", "-D", "PREROUTING",
				"-d", hostIP, "-j", "MARK", "--set-mark", strconv.Itoa(ids.DownloadID))
		} else { // Was blocked
			// Remove iptables DROP rule
			executeCommand("iptables", "-t", "filter", "-D", "FORWARD",
				"-d", hostIP, "-j", "DROP")
		}
	}

	// Remove from tracking
	delete(tl.hostDict, hostID)

	return nil
}
