package shell

import (
	"fmt"
	"strconv"
)

func (s *ShellSession) Spoof(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Usage: spoof <list|stop> [host_id]")
		return
	}

	switch args[0] {
	case "list":
		fmt.Println("📋 Current spoofing sessions:")
		s.store.DisplaySpoofList()
	case "stop":
		if len(args) < 2 {
			fmt.Println("❌ Usage: spoof stop <host_id>")
			return
		}
		hostId, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("❌ Invalid host ID '%s': must be a number\n", args[1])
			return
		}
		if hostId < 0 {
			fmt.Println("❌ Host ID must be a positive number")
			return
		}

		s.store.SpoofManager.Stop(int64(hostId))
		fmt.Printf("✅ Spoofing stopped for host ID: %d\n", hostId)
	default:
		fmt.Printf("❌ Unknown spoof command: '%s'\n", args[0])
		fmt.Println("💡 Available options: list, stop")
	}
}
