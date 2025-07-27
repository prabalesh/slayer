package shell

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/prabalesh/slayer/internal/limiter"
	"github.com/prabalesh/slayer/internal/store"
	"github.com/prabalesh/slayer/internal/utils/color"
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

func (s *ShellSession) CheckRequirements() error {
	fmt.Print("🔍 Checking system requirements for Slayer...\n\n")

	allPassed := true

	// OS Check
	if runtime.GOOS == "linux" {
		fmt.Println(color.GreenText("✅ Linux OS detected", false))
	} else {
		fmt.Printf(color.RedText("❌ Slayer requires Linux (detected: %s)\n", false), runtime.GOOS)
		allPassed = false
	}

	// Root Check
	if os.Geteuid() == 0 {
		fmt.Println(color.GreenText("✅ Running as root", false))
	} else {
		fmt.Println(color.RedText("❌ Slayer must be run as root (try with sudo)", false))
		allPassed = false
	}

	// Required binaries
	requiredTools := []string{"iptables", "tc", "ip"}

	for _, tool := range requiredTools {
		if path, err := exec.LookPath(tool); err == nil {
			fmt.Print(color.GreenText(fmt.Sprintf("✅ %s found at %s\n", tool, path), false))
		} else {
			fmt.Print(color.RedText(fmt.Sprintf("❌ %s not found in PATH\n", tool), false))
			allPassed = false
		}
	}

	fmt.Println()

	if !allPassed {
		return fmt.Errorf("❌ One or more requirements are not met. Slayer cannot start.")
	}

	fmt.Print(color.GreenText("🎉 All requirements passed! You're good to go.\n\n", true))
	return nil
}

func (s *ShellSession) Start() {
	s.setupSignalHandler()
	s.printWelcome()

	if err := s.CheckRequirements(); err != nil {
		fmt.Println(color.RedText(err.Error(), true))
		os.Exit(1)
	}

	for {
		input := s.readInput()
		if input == "" {
			continue
		}

		command, args := s.parseCommand(input)

		if s.shouldExit(command) {
			s.Close()
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
		s.Close()
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
	fmt.Print(color.BlueText("⚡ slayer> ", false))
	input, err := s.reader.ReadString('\n')
	if err != nil {
		fmt.Print(color.RedText(fmt.Sprintf("❌ Error reading input: %v\n", err), false))
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
		s.Limit(args)
	case "unlimit":
		s.Unlimit(args)
	case "spoof":
		s.Spoof(args)
	case "clear":
		fmt.Print("\033[2J\033[H")
	default:
		fmt.Printf("❌ Unknown command: '%s'\n", command)
		fmt.Println("💡 Type 'help' to see available commands")
	}
}

func (s *ShellSession) Close() {
	for _, host := range s.store.Hosts {
		if host.Limited {
			fmt.Printf("Removing limit on %s...\n", host.IP.String())
			err := limiter.Remove(host.IP.String(), s.store.Iface.Name)
			if err != nil {
				fmt.Printf("Can't remove limit on %s\n", host.IP.String())
				return
			}
			fmt.Printf("Removed limit on %s\n", host.IP.String())
		}
	}
}
