package networking

import (
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/mdlayher/arp"
)

type Job func()

type Pool struct {
	workerQueue chan Job
	wg          sync.WaitGroup
}

func NewPool(workerCount int) *Pool {
	pool := &Pool{workerQueue: make(chan Job)}
	pool.wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer pool.wg.Done()
			for job := range pool.workerQueue {
				job()
			}
		}()
	}

	return pool
}

func (p *Pool) AddJob(job Job) {
	p.workerQueue <- job
}

func (p *Pool) Wait() {
	close(p.workerQueue)
	p.wg.Wait()
}

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
	// Check if we got valid inputs
	if a.iface == nil || len(ips) == 0 {
		return []Host{}
	}

	var activeHosts []Host
	var hostsMutex sync.Mutex

	pool := NewPool(a.maxWorkers)

	for _, ip := range ips {
		currentIP := ip
		job := func() {
			host := a.scanSingleIP(currentIP)
			if host != nil {
				// We found a device! Add it to our list safely
				hostsMutex.Lock()
				activeHosts = append(activeHosts, *host)
				hostsMutex.Unlock()
			}
		}
		pool.AddJob(job)

	}
	pool.Wait()
	return activeHosts
}

// scanSingleIP tries to find a device at one specific IP address
func (a *ArpScanner) scanSingleIP(ip net.IP) *Host {
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

	names, err := net.LookupAddr(ip.String())
	if err == nil && len(names) > 0 {
		host.Name = names[0]
	}

	return host
}
