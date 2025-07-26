package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prabalesh/slayer/internal/networking"
)

var activeHosts []networking.Host

func printActiveHosts() {
	if len(activeHosts) <= 0 {
		fmt.Println("No devices online")
		return
	}

	fmt.Print("ID\tIP Address\tMAC Addres\t\tHostname\n")
	for _, host := range activeHosts {
		fmt.Printf("%d\t%s\t%s\t%s\n", host.Id, host.IP, host.MAC, host.Name)
	}
}

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
	activeHosts = arpScanner.Scan(ips)
	timeTaken := time.Since(startedTime)
	printActiveHosts()

	fmt.Printf("Scan completed in %v\n", timeTaken)

}

func help() {
	fmt.Println("scan - scanning for all the host in the network")
	fmt.Println("list - lists all the active devices in the network")
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
		case "list":
			printActiveHosts()
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
