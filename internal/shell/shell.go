package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/prabalesh/slayer/internal/store"
)

type ShellSession struct {
	store store.Store
}

func NewShell(s store.Store) *ShellSession {
	return &ShellSession{store: s}
}

func (s *ShellSession) Start() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		inputSlice := strings.Split(input, " ")
		command := inputSlice[0]
		args := inputSlice[1:]

		if command == "quit" || command == "exit" {
			break
		}
		switch command {
		case "scan":
			s.RunNetworkScan()
		case "list":
			s.DisplayActiveHosts()
		case "help":
			s.Help()
		case "limit":
			s.Limit(args)
		default:
			fmt.Println("Invalid command use help to list all the commands")
		}
	}
}
