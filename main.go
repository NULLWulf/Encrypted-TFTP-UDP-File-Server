package main

import "CSC445_Assignment2/tftp"

func main() {
	// Parse the command line arguments
	parseProgramArguments()
	// Run the application in the specified mode
	tftp.Test{}.Test()
	if Mode == "server" {
		//RunServerMode()
	} else {
		RunClientMode()
	}
}
