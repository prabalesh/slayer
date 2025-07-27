package shell

import "fmt"

func (s *ShellSession) Help() {
	fmt.Println("scan - scanning for all the host in the network")
	fmt.Println("list - lists all the active devices in the network")
	fmt.Println("quit | exit - for existing")
}
