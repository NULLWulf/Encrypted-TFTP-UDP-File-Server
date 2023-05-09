package main

import (
	"log"
)

func main() {
	// Parse the command line arguments
	//AESTester()
	parseProgramArguments()
	// Run the application in the specified mode

	switch Mode {
	case "server":
		log.Printf("Running in server mode on port %d with window size %d\n", Port, WindowSize)
		RunServerMode()

	case "client":
		RunClientMode()

	default:
		client, err := NewTFTPClient() // instantiate a new TFTP client
		if err != nil {
			log.Printf("Error Creating TFTP Client: %s\n", err)
			return
		}
		defer client.Close()
		_, _, _ = client.RequestFile("https://rare-gallery.com/uploads/posts/577429-star-wars-high.jpg") // request the file via url
	}
}
