package store

import (
	"context"
	"net"
	"sync"

	"github.com/prabalesh/slayer/internal/limiter"
)

// SpoofManager controls spoofing operations per host.
type SpoofManager struct {
	cancelMap map[int64]context.CancelFunc
	mu        sync.Mutex
}

// Host represents a discovered device on the network.
type Host struct {
	ID            int64            // Unique identifier
	IP            net.IP           // IPv4 address
	MAC           net.HardwareAddr // MAC address
	Hostname      string           // Resolved hostname (if any)
	Online        bool             // Whether host is currently reachable
	Limited       bool             // Whether traffic is currently throttled
	UploadSpeed   string
	DownloadSpeed string
}

// Store holds global network context and all known hosts.
type Store struct {
	Iface        *net.Interface   // Active network interface
	GatewayIP    net.IP           // Default gateway IP
	GatewayMAC   net.HardwareAddr // Default gateway MAC
	CIDR         string           // CIDR of the interface (e.g. 192.168.1.0/24)
	Hosts        map[int64]*Host  // Keyed by IP string (e.g. "192.168.1.101")
	SpoofManager *SpoofManager
	Limiter      *limiter.Limiter
}
