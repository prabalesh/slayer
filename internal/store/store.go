// Package store initializes and manages network store data.
package store

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/prabalesh/slayer/internal/limiter"
	"github.com/prabalesh/slayer/internal/networking"
	"github.com/prabalesh/slayer/internal/spoof"
)

// NewStore creates and returns a fully initialized Store.
func NewStore() (*Store, error) {
	iface, err := networking.GetActiveWiFiInterface()
	if err != nil {
		return nil, fmt.Errorf("failed to get active interface: %w", err)
	}

	gatewayIP, err := networking.GetDefaultGatewayIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway IP: %w", err)
	}

	gatewayMAC, err := networking.GetGatewayMAC(gatewayIP)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway MAC address: %w", err)
	}

	cidr, err := networking.GetInterfaceCIDR(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface CIDR: %w", err)
	}

	newLimiter := limiter.NewLimiter()
	newLimiter.Init(iface)

	store := &Store{
		Iface:        iface,
		GatewayIP:    gatewayIP,
		GatewayMAC:   gatewayMAC,
		CIDR:         cidr,
		Hosts:        make(map[int64]*Host),
		SpoofManager: NewSpoofManager(),
		Limiter:      newLimiter,
	}

	return store, nil
}

// AddHost adds a new host or updates an existing one in the store.
func (s *Store) AddHost(host *Host) {
	if host == nil || host.IP == nil {
		return
	}
	s.Hosts[host.ID] = host
}

// GetHost retrieves a host by IP address string.
func (s *Store) GetHost(hostId int64) (*Host, bool) {
	host, exists := s.Hosts[hostId]
	return host, exists
}

// ListHosts returns all discovered hosts.
func (s *Store) ListHosts() []*Host {
	hosts := make([]*Host, 0, len(s.Hosts))
	for _, host := range s.Hosts {
		hosts = append(hosts, host)
	}
	return hosts
}

// spoofmanager

// NewSpoofManager returns a new instance of SpoofManager.
func NewSpoofManager() *SpoofManager {
	return &SpoofManager{
		cancelMap: make(map[int64]context.CancelFunc),
	}
}

// Start begins spoofing the specified host.
func (sm *SpoofManager) Start(host *Host, iface *net.Interface, gatewayIP net.IP, gatewayMAC net.HardwareAddr) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.cancelMap[host.ID]; exists {
		return // already spoofing
	}

	ctx, cancel := context.WithCancel(context.Background())
	sm.cancelMap[host.ID] = cancel

	go spoof.Spoof(ctx, iface, host.IP, host.MAC, gatewayIP, gatewayMAC)
	time.Sleep(1 * time.Second)
}

// Stop ends spoofing for a specific host.
func (sm *SpoofManager) Stop(hostID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if cancel, exists := sm.cancelMap[hostID]; exists {
		cancel()
		delete(sm.cancelMap, hostID)
	}
}

// StopAll stops spoofing for all hosts.
func (sm *SpoofManager) StopAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, cancel := range sm.cancelMap {
		cancel()
	}
	sm.cancelMap = make(map[int64]context.CancelFunc)
}

func (s *Store) DisplaySpoofList() {
	for id, _ := range s.SpoofManager.cancelMap {
		host := s.Hosts[id]
		fmt.Printf("%d\t%s\t%s\n", host.ID, host.IP.String(), host.Hostname)
	}
}
