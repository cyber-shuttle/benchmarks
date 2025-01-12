package main

import (
	"flag"
	router "grpcsh/router"
	"log"
)

func main() {

	// argument parsing
	routerUrl := flag.String("r", "localhost:50051", "Router URL")
	flag.Parse()

	// validation
	if *routerUrl == "" {
		log.Fatalf("Router URL must be provided using -r")
	}

	// logic
	log.Println("Router URL:", *routerUrl)
	router.Start(*routerUrl)
}
