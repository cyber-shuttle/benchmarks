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
		if err := stream.Send(&pb.Message{To: *peerId, Flag: pb.Flag_COMMAND, Data: []byte(*command)}); err != nil {
			errChan <- fmt.Errorf("Error sending command: %v", err)
			return
		}
		// send stdin if present
		stat, err := os.Stdin.Stat()
		if err != nil {
			errChan <- fmt.Errorf("error stating stdin: %v", err)
			return
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			errChan <- nil
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
				break
			}
			if err := stream.Send(&pb.Message{To: *peerId, Flag: pb.Flag_MSG_STDIN, Data: buf[:n]}); err != nil {
				errChan <- fmt.Errorf("error sending stdin to stream: %v", err)
				break
			}
		}
		if err := stream.Send(&pb.Message{To: *peerId, Flag: pb.Flag_EOF_STDIN}); err != nil {
			log.Printf("error sending EOF stdin to stream: %v", err)
		}
		os.Stdin.Close()
	}()

	// inbound stream
	go func() {
		eof_stdout := false
		eof_stderr := false
		for {
			if eof_stdout && eof_stderr {
				errChan <- nil
				return
			}
			result, err := stream.Recv()
			if err != nil {
				errChan <- fmt.Errorf("error receiving from stream: %v", err)
				return
			}
			// Write to appropriate output
			switch result.Flag {
			case pb.Flag_MSG_STDERR:
				if _, err := os.Stderr.Write(result.Data); err != nil {
					errChan <- fmt.Errorf("error writing to stderr: %v", err)
					return
				}
			case pb.Flag_MSG_STDOUT:
				if _, err := os.Stdout.Write(result.Data); err != nil {
					errChan <- fmt.Errorf("error writing to stdout: %v", err)
					return
				}
			case pb.Flag_EOF_STDERR:
				eof_stderr = true
				os.Stderr.Close()
			case pb.Flag_EOF_STDOUT:
				eof_stdout = true
				os.Stdout.Close()
			}
		}
	}()

	// Wait for both goroutines to complete
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			log.Fatal(err)
		}
	}

	stream.CloseSend()
}
