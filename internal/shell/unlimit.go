package shell

import (
	"fmt"
	"strconv"

	"github.com/prabalesh/slayer/internal/limiter"
)

func (s *ShellSession) Unlimit(args []string) {
	if len(args) < 1 {
		fmt.Printf("Don't enough arguments")
	}
	// Parse host ID
	hostId, err := strconv.Atoi(args[0])
	if err != nil || hostId < 0 {
		fmt.Printf("❌ Invalid host ID '%s': must be a positive number\n", args[0])
		return
	}
	host := s.store.Hosts[int64(hostId)]

	err = limiter.Remove(host.IP.String(), s.store.Iface.Name)
	if err != nil {
		fmt.Printf("Can't unlimit host %s: %s\n", host.IP, err)
		return
	}
	s.store.Hosts[int64(hostId)].Limited = false
	s.store.SpoofManager.Stop(int64(hostId))
	fmt.Printf("Unlimited host %s\n", host.IP)
}
