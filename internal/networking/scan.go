package networking

import (
	"context"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/mdlayher/arp"
)

type Host struct {
	IP   net.IP
	MAC  string
	Name string
}

type ArpScanner struct {
	iface      *net.Interface
	timeout    time.Duration
	maxWorkers int
}

func NewArpScanner(iface *net.Interface) *ArpScanner {
	return &ArpScanner{
		iface:      iface,
		timeout:    2 * time.Second,
		maxWorkers: 64,
	}
}

func (a *ArpScanner) Scan(ips []net.IP) []Host {
	ctx := context.Background()

	// Check if we got valid inputs
	if a.iface == nil || len(ips) == 0 {
		return []Host{}
	}

	var activeHosts []Host
	var hostsMutex sync.Mutex

	// Create a channel to limit how many goroutines run at once
	// Think of it like a parking lot with limited spaces
	workerSemaphore := make(chan struct{}, a.maxWorkers)

	var wg sync.WaitGroup

	for _, ip := range ips {
		currentIP := ip
		wg.Add(1)
		go func(ipToScan net.IP) {
			defer wg.Done()

			// Try to get a "parking space" in our semaphore
			select {
			case workerSemaphore <- struct{}{}: // Got a space!
				defer func() { <-workerSemaphore }() // Give up the space when done
			case <-ctx.Done(): // Someone cancelled us
				return
			}

			// Try to find the device at this IP address
			host := a.scanSingleIP(ctx, ipToScan)
			if host != nil {
				// We found a device! Add it to our list safely
				hostsMutex.Lock()
				activeHosts = append(activeHosts, *host)
				hostsMutex.Unlock()
			}
		}(currentIP)
	}
	wg.Wait()
	return activeHosts
}

// scanSingleIP tries to find a device at one specific IP address
func (a *ArpScanner) scanSingleIP(ctx context.Context, ip net.IP) *Host {
	// Create a connection to send ARP requests
	conn, err := arp.Dial(a.iface)
	if err != nil {
		// Can't create connection, maybe interface is down?
		return nil
	}
	defer conn.Close() // Always close the connection when we're done

	// Set how long to wait for a response
	deadline := time.Now().Add(a.timeout)
	conn.SetDeadline(deadline)

	// Convert IP to the format the ARP library wants
	ipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		// IP address is in wrong format
		return nil
	}

	// Send ARP request: "Hey, who has this IP address?"
	mac, err := conn.Resolve(ipAddr)
	if err != nil {
		// No response = no device at this IP
		return nil
	}

	// We found a device! Create a Host record
	host := &Host{
		IP:  ip,
		MAC: mac.String(),
	}

	// Try to find the device's name (like "Johns-iPhone")
	// This is optional and might be slow, so we do it in a separate step
	select {
	case <-ctx.Done():
		// Someone cancelled us, return what we have
		return host
	default:
		// Try to look up the hostname
		names, err := net.LookupAddr(ip.String())
		if err == nil && len(names) > 0 {
			host.Name = names[0]
		}
	}

	return host
}
