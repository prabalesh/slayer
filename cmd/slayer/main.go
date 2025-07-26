package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prabalesh/slayer/internal/networking"
)

func runScan() {
	iface, err := networking.GetActiveWiFiInterface()
	if err != nil {
		log.Fatal("Error detecting interface:", err)
	}

	cidr, err := networking.GetInterfaceCIDR(iface)
	if err != nil {
		log.Fatal("Error getting CIDR:", err)
	}

	fmt.Printf("Detected interface: %s\n", iface.Name)
	fmt.Printf("Detected CIDR: %s\n", cidr)

	ips, err := networking.GenerateIPsFromCIDR(cidr)
	if err != nil {
		log.Fatal(err)
	}

	startedTime := time.Now()
	arpScanner := networking.NewArpScanner(iface)
	hosts := arpScanner.Scan(ips)
	timeTaken := time.Since(startedTime)
	for _, host := range hosts {
		fmt.Println(host)
	}
	fmt.Printf("Scan completed in %v\n", timeTaken)

}

func help() {
	fmt.Println("scan - scanning for all the host in the network")
	fmt.Println("quit | exit - for existing")
}

func shell() {
	for {
		var input string
		fmt.Print("> ")
		fmt.Scanf("%s\n", &input)

		inputSlice := strings.Split(input, " ")
		command := inputSlice[0]
		// args := inputSlice[1:]

		if command == "quit" || command == "exit" {
			break
		}
		switch command {
		case "scan":
			runScan()
		case "help":
			help()
		default:
			fmt.Println("Invalid command use help to list all the commands")
		}
	}
}

func main() {
	shell()
}
