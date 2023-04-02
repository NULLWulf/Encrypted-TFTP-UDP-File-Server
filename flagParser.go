package main

import (
	"flag"
	"log"
)

var (
	Address string
	Mode    string
	Port    int
	DropPax bool
)

// parseProgramArguments parses the command line arguments and sets the global variables based on them
// if configuration is valid the program will continue, otherwise it will exit with an error code
// contains options for server, client, address, simulated packet drops.
func parseProgramArguments() {
	flag.StringVar(&Mode, "Mode", "", "Application mode: 'server' or client'.")
	flag.StringVar(&Address, "Address", "", "Remote address to connect to while in Client mode, this field is ignored when set in server mode.")
	flag.IntVar(&Port, "Port", 7500, "Port the application will listen to while in server mode.")
	flag.BoolVar(&DropPax, "DropPax", false, "Simulate dropping packets while in server mode.")
	flag.Parse()

	if Mode != "server" && Mode != "client" {
		log.Fatalf("Invalid Mode Specified.  Must be use either '-Mode server' or '-Mode client'.")
	}

	if Mode == "server" && Address != "" {
		log.Println("Warning: Address argument is ignored when application set to server mode.")
	}

	if Mode == "client" && Address == "" {
		log.Fatalf("Invalid Address.  Address must be specified for client mode.")
	}

	if DropPax && Mode == "client" {
		log.Println("Warning: DropPax argument is ignored when application set to client mode.")
	}

	if DropPax && Mode == "server" {
		log.Println("Application set to server mode with simulated dropped packets..")
	}
}
