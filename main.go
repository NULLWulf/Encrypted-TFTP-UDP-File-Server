package main

import (
	"log"
)

var IQ *ImageQueueObj

func main() {

	//TestRetrievingFile()
	//os.Exit(0)
	// Parse the command line arguments
	parseProgramArguments()
	// Run the application in the specified mode
	if Mode == "server" {
		IQ = NewImageQueueObj()
		log.Printf("Running in server mode on port %d with window size %d\n", Port, WindowSize)
		RunServerMode()
	} else {
		log.Printf("Running in client mode\n")
		RunClientMode()
	}
}
