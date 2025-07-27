package shell

import "fmt"

func (s *ShellSession) DisplayActiveHosts() {
	if len(s.store.Hosts) <= 0 {
		fmt.Println("No devices online")
		return
	}

	fmt.Print("ID\tIP Address\tMAC Addres\t\tHostname\t\tLimited\n")
	for id, host := range s.store.Hosts {
		fmt.Printf("%d\t%s\t%s\t%s\t\t%v\n", id, host.IP, host.MAC, host.Hostname, host.Limited)
	}
}
