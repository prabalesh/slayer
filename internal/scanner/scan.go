package scanner

import (
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mdlayher/arp"
	"github.com/prabalesh/slayer/internal/store"
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

// FastConsistentArpScanner - Fast but consistent results
type ArpScanner struct {
	idCounter  int64
	iface      *net.Interface
	timeout    time.Duration
	maxWorkers int
	store      *store.Store
}

func NewArpScanner(s *store.Store) *ArpScanner {
	return &ArpScanner{
		iface:      s.Iface,
		timeout:    2 * time.Second, // Back to your original 2s
		maxWorkers: 50,              // Back to your original 50
		store:      s,
	}
}

// Simple scan with just one retry for failed IPs
func (a *ArpScanner) Scan(ips []net.IP) {
	if a.iface == nil || len(ips) == 0 {
		return
	}

	var hostsMutex sync.Mutex
	var foundIPs sync.Map // Thread-safe map to track found IPs

	// Round 1: Main scan
	pool := NewPool(a.maxWorkers)
	for _, ip := range ips {
		currentIP := ip
		job := func() {
			host := a.scanSingleIP(currentIP)
			if host != nil {
				foundIPs.Store(currentIP.String(), true)
				hostsMutex.Lock()
				a.store.AddHost(host)
				hostsMutex.Unlock()
			}
		}
		pool.AddJob(job)
	}
	pool.Wait()

	// Quick retry for missed IPs (only takes a few seconds)
	var missedIPs []net.IP
	for _, ip := range ips {
		if _, found := foundIPs.Load(ip.String()); !found {
			missedIPs = append(missedIPs, ip)
		}
	}

	// Only retry if we missed some and it's reasonable amount
	if len(missedIPs) > 0 && len(missedIPs) < len(ips)/2 {
		time.Sleep(200 * time.Millisecond) // Short pause

		retryPool := NewPool(a.maxWorkers)
		for _, ip := range missedIPs {
			currentIP := ip
			job := func() {
				host := a.scanSingleIPWithRetry(currentIP)
				if host != nil {
					hostsMutex.Lock()
					a.store.AddHost(host)
					hostsMutex.Unlock()
				}
			}
			retryPool.AddJob(job)
		}
		retryPool.Wait()
	}
}

// Your original scan method (proven to work)
func (a *ArpScanner) scanSingleIP(ip net.IP) *store.Host {
	conn, err := arp.Dial(a.iface)
	if err != nil {
		return nil
	}
	defer conn.Close()

	deadline := time.Now().Add(a.timeout)
	conn.SetDeadline(deadline)

	ipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return nil
	}

	mac, err := conn.Resolve(ipAddr)
	if err != nil {
		return nil
	}

	id := atomic.AddInt64(&a.idCounter, 1)
	host := &store.Host{
		ID:  id,
		IP:  ip,
		MAC: mac,
	}

	// Background hostname lookup
	go func() {
		if names, err := net.LookupAddr(ip.String()); err == nil && len(names) > 0 {
			host.Hostname = names[0]
		}
	}()

	return host
}

// Single retry with slightly longer timeout
func (a *ArpScanner) scanSingleIPWithRetry(ip net.IP) *store.Host {
	conn, err := arp.Dial(a.iface)
	if err != nil {
		return nil
	}
	defer conn.Close()

	// Slightly longer timeout for retry
	deadline := time.Now().Add(a.timeout + 500*time.Millisecond)
	conn.SetDeadline(deadline)

	ipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return nil
	}

	mac, err := conn.Resolve(ipAddr)
	if err != nil {
		return nil
	}

	id := atomic.AddInt64(&a.idCounter, 1)
	host := &store.Host{
		ID:  id,
		IP:  ip,
		MAC: mac,
	}

	// Background hostname lookup
	go func() {
		if names, err := net.LookupAddr(ip.String()); err == nil && len(names) > 0 {
			host.Hostname = names[0]
		}
	}()

	return host
}
