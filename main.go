package main

func main() {
	// Parse the command line arguments
	parseProgramArguments()
	// Run the application in the specified mode
	if Mode == "server" {
		//RunServerMode()
	} else {
		RunClientMode()
	}
}
