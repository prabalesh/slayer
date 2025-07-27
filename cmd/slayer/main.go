package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prabalesh/slayer/internal/networking"
	"github.com/prabalesh/slayer/internal/networking/spoof"
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

func limit_traffic(args []string) {
	// parse args
	fmt.Println(args)
	hostId, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid argument")
		return
	}
	var targetHost networking.Host
	for _, host := range activeHosts {
		if host.Id == uint64(hostId) {
			targetHost = host
			break
		}
	}

	// get the default gateway ip and mac addr
	gatewayIp, err := networking.GetDefaultGatewayIP()
	if err != nil {
		fmt.Println("Can't get the gateway ip")
		return
	}
	fmt.Printf("Gateway IP: %s\n", gatewayIp)
	gatewayMAC, err := networking.GetGatewayMAC(gatewayIp)
	if err != nil {
		fmt.Printf("Can't get gateway mac: %s\n", err)
		return
	}
	fmt.Printf("Gateway MAC address %s\n", gatewayMAC)
	iface, err := networking.GetActiveWiFiInterface()
	if err != nil {
		log.Fatal("Error detecting interface:", err)
	}
	fmt.Printf("Detected interface: %s\n", iface.Name)

	// start arp spoofing
	// go func() {
	spoof.Spoof(iface, targetHost.IP, net.HardwareAddr(targetHost.MAC), gatewayIp, gatewayMAC)
	// }()

	// label the network

	// reduce the traffic
}

func help() {
	fmt.Println("scan - scanning for all the host in the network")
	fmt.Println("list - lists all the active devices in the network")
	fmt.Println("quit | exit - for existing")
}

func shell() {
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
			runScan()
		case "list":
			printActiveHosts()
		case "help":
			help()
		case "limit":
			limit_traffic(args)
		default:
			fmt.Println("Invalid command use help to list all the commands")
		}
	}
}

func main() {
	shell()
}
