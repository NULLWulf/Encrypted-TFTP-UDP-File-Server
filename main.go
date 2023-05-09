package main

import (
	"log"
	"os"
)

func main() {
	AESTester()
	os.Exit(0)
	//for i := 0; i < 100; i++ {
	//	DHKETester()
	//}

	//TestRetrievingFile()
	//os.Exit(0)
	// Parse the command line arguments
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
