package main

import (
	"context"
	"flag"
	"fmt"
	pb "grpcsh/pb"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	// argument parsing
	peerId := flag.String("i", "local", "The peer to run the command on")
	sockPath := flag.String("s", "agent.sock", "The socket to connect to")
	command := flag.String("c", "", "The command to execute")
	flag.Parse()

	// validation
	if *peerId == "" {
		log.Fatal("Peer ID must be provided using -i")
	}
	if *sockPath == "" {
		log.Fatal("Socket path must be provided using -s")
	}
	if *command == "" {
		log.Fatal("Shell command must be provided using -c")
	}

	// logic
	conn, err := grpc.NewClient("unix://"+*sockPath, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewExecutorServiceClient(conn)

	// Create a context
	ctx := context.Background()
	stream, err := client.Exec(ctx)
	if err != nil {
		log.Fatalf("Error creating stream: %v", err)
	}

	// Create error channel for goroutines
	errChan := make(chan error, 2)

	// outbound stream
	go func() {
		// send the command
		if err := stream.Send(&pb.Message{Peer: *peerId, Data: &pb.Message_Command{Command: *command}}); err != nil {
			log.Fatalf("Error sending command: %v", err)
		}
		// send stdin if present
		stat, err := os.Stdin.Stat()
		if err != nil {
			errChan <- fmt.Errorf("error stating stdin: %v", err)
			stream.CloseSend()
			return
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			errChan <- nil
			stream.CloseSend()
			return
		}
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if err != io.EOF {
					errChan <- fmt.Errorf("error reading stdin: %v", err)
				} else {
					errChan <- nil
				}
				stream.CloseSend()
				return
			}
			if err := stream.Send(&pb.Message{Peer: *peerId, Data: &pb.Message_Stdin{Stdin: buf[:n]}}); err != nil {
				errChan <- fmt.Errorf("error sending stdin to stream: %v", err)
				return
			}
		}
	}()

	// inbound stream
	go func() {
		for {
			result, err := stream.Recv()
			if err == io.EOF {
				errChan <- nil
				return
			}
			if err != nil {
				errChan <- fmt.Errorf("error receiving from stream: %v", err)
				return
			}

			// Write to appropriate output
			switch result.Data.(type) {
			case *pb.Result_Stderr:
				if _, err := os.Stderr.Write(result.Data.(*pb.Result_Stderr).Stderr); err != nil {
					errChan <- fmt.Errorf("error writing to stderr: %v", err)
					return
				}
			case *pb.Result_Stdout:
				if _, err := os.Stdout.Write(result.Data.(*pb.Result_Stdout).Stdout); err != nil {
					errChan <- fmt.Errorf("error writing to stdout: %v", err)
					return
				}
			}
		}
	}()

	// Wait for both goroutines to complete
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			log.Fatal(err)
		}
	}
}
