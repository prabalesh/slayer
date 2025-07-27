package shell

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/prabalesh/slayer/internal/store"
)

type ShellSession struct {
	store  store.Store
	reader *bufio.Reader
}

func NewShell(s store.Store) *ShellSession {
	return &ShellSession{
		store:  s,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (s *ShellSession) Start() {
	s.setupSignalHandler()
	s.printWelcome()

	for {
		input := s.readInput()
		if input == "" {
			continue
		}

		command, args := s.parseCommand(input)

		if s.shouldExit(command) {
			fmt.Println("\n🔴 Shutting down Slayer...")
			break
		}

		s.executeCommand(command, args)
	}
}

func (s *ShellSession) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\n🔴 Interrupted! Exiting Slayer...")
		os.Exit(0)
	}()
}

func (s *ShellSession) printWelcome() {
	fmt.Print(`
 ╔═══════════════════════════════════════════════════╗
 ║  ███████ ██       █████  ██    ██ ███████ ██████   ║
 ║  ██      ██      ██   ██  ██  ██  ██      ██   ██  ║
 ║  ███████ ██      ███████   ████   █████   ██████   ║
 ║       ██ ██      ██   ██    ██    ██      ██   ██  ║
 ║  ███████ ███████ ██   ██    ██    ███████ ██   ██  ║
 ╚═══════════════════════════════════════════════════╝

    🔥 Network Security Testing Tool 🔥

 📝 Type 'help' for commands | Type 'quit' to exit
`)
}

func (s *ShellSession) readInput() string {
	fmt.Print("⚡ slayer> ")
	input, err := s.reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Error reading input: %v\n", err)
		return ""
	}
	return strings.TrimSpace(input)
}

func (s *ShellSession) parseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func (s *ShellSession) shouldExit(command string) bool {
	return command == "quit" || command == "exit"
}

func (s *ShellSession) executeCommand(command string, args []string) {
	switch command {
	case "scan":
		fmt.Println("🔍 Initiating network scan...")
		s.RunNetworkScan()
	case "list":
		fmt.Println("📋 Displaying active hosts...")
		s.DisplayActiveHosts()
	case "help":
		s.HandleHelp()
	case "limit":
		if len(args) < 2 {
			fmt.Println("❌ Usage: limit <target_ip> <bandwidth_limit>")
			return
		}
		fmt.Printf("⚠️  Setting bandwidth limit for %s to %s...\n", args[0], args[1])
		s.Limit(args)
	case "spoof":
		if len(args) < 2 {
			fmt.Println("❌ Usage: spoof <target_ip> <gateway_ip>")
			return
		}
		fmt.Printf("🎭 Starting ARP spoofing: %s -> %s...\n", args[0], args[1])
		s.Spoof(args)
	case "clear":
		fmt.Print("\033[2J\033[H")
	default:
		fmt.Printf("❌ Unknown command: '%s'\n", command)
		fmt.Println("💡 Type 'help' to see available commands")
	}
}
