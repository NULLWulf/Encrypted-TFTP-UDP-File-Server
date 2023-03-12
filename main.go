package main

import (
	"log"
)

var IQ *ImageQueueObj

func main() {
	// Parse the command line arguments
	parseProgramArguments()
	// Run the application in the specified mode
	//tftp.Test{}.Test()
	if Mode == "server" {
		IQ = NewImageQueueObj()
		log.Printf("Running in server mode\n")
		RunServerMode()
	} else {
		log.Printf("Running in client mode\n")
		RunClientMode()
	}
}
