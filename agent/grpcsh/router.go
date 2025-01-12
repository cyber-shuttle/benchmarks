package main

import (
	"flag"
	"fmt"
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
	fmt.Println("Router URL:", *routerUrl)
	router.Start(*routerUrl)
}
