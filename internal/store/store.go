// Package store initializes and manages network store data.
package store

import (
	"fmt"

	"github.com/prabalesh/slayer/internal/networking"
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

	store := &Store{
		Iface:      iface,
		GatewayIP:  gatewayIP,
		GatewayMAC: gatewayMAC,
		CIDR:       cidr,
		Hosts:      make(map[int64]*Host),
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
