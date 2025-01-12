package main

import (
	"flag"
	agent "grpcsh/agent"
	"log"
	"os"
)

func main() {

	// argument parsing
	peerID := flag.String("i", "", "Peer ID")
	routerUrl := flag.String("r", "localhost:50051", "Router URL")
	socketPath := flag.String("s", "agent.sock", "Socket Path")
	flag.Parse()

	// validation
	if *peerID == "" {
		log.Fatalf("Peer ID must be provided using -i")
	}
	if *routerUrl == "" {
		log.Fatalf("Router URL must be provided using -r")
	}
	if *socketPath == "" {
		log.Fatalf("Socket Path must be provided using -s")
	}
	if err := os.RemoveAll(*socketPath); err != nil {
		log.Fatalf("Failed to remove existing socket: %v", err)
	}

	// logic
	log.Println("Peer ID:", *peerID)
	log.Println("Router URL:", *routerUrl)
	log.Println("Socket Path:", *socketPath)
	agent.Start(*peerID, *routerUrl, *socketPath)
}
