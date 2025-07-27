package shell

import "fmt"

var commands = map[string]string{
	"scan":  "Perform network scan to discover active hosts",
	"list":  "Display all discovered active hosts",
	"limit": "Set bandwidth limits on target hosts",
	"spoof": "Perform ARP spoofing attack",
	"help":  "Show available commands",
	"quit":  "Exit Slayer",
	"exit":  "Exit Slayer",
	"clear": "Clear the terminal screen",
}

func (s *ShellSession) HandleHelp() {
	fmt.Println(`
══════════════════════════════════════════════════════════════
                        🔥 SLAYER COMMANDS 🔥
══════════════════════════════════════════════════════════════`)

	commandOrder := []string{"scan", "list", "limit", "spoof", "clear", "help", "quit"}

	for _, cmd := range commandOrder {
		if desc, exists := commands[cmd]; exists {
			fmt.Printf("║  %-8s │ %-45s ║\n", cmd, desc)
		}
	}

	fmt.Println(`══════════════════════════════════════════════════════════════`)
	fmt.Println()
}
