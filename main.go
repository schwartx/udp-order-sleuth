package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Define global variables to store command-line arguments
var (
	sendFlag     = flag.Bool("send", false, "Start as a sender")
	address      = flag.String("addr", "225.0.0.250:5001", "Multicast address")
	revFlag      = flag.Bool("rev", false, "Start as a receiver")
	sendInterval = flag.Duration("interval", 1*time.Second, "Send interval for sender")
)

func main() {
	// Parse command-line arguments
	flag.Parse()

	// Check if both -send and -rev are set or neither is set
	if (*sendFlag && *revFlag) || (!*sendFlag && !*revFlag) {
		printHelpAndExit()
	}

	// Call the appropriate function based on the flag
	if *sendFlag {
		sender(*sendInterval)
	} else if *revFlag {
		receiver()
	}
}

func printHelpAndExit() {
	fmt.Println("Usage: udp-order-sleuth [OPTIONS]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults() // Print default values for all defined flags
	os.Exit(1)
}
