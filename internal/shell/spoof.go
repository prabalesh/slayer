package shell

import (
	"fmt"
	"strconv"
)

func (s *ShellSession) Spoof(args []string) {
	switch args[0] {
	case "list":
		s.store.DisplaySpoofList()
	case "stop":
		if len(args) > 2 {
			fmt.Println("No host id found")
			return
		}
		hostId, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Host id should be a number")
		}
		s.store.SpoofManager.Stop(int64(hostId))
		fmt.Println("Spoofing stopped")
	}
}
