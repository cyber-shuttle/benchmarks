package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"

	pb "grpcsh/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type executorServer struct {
	pb.UnimplementedExecutorServiceServer
}

// global variables
var bus *Router
var channelSvcClient pb.ChannelServiceClient

func (s *executorServer) Exec(stream pb.ExecutorService_ExecServer) error {
	command, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive command: %w", err)
	}
	// command should always be command
	peer := command.Peer
	if peer == "local" {
		go execLocal(stream, command)
	} else {
		channel, err := channelSvcClient.CreateChannel(context.Background(), &emptypb.Empty{})
		if err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}
		ch := bus.Channel(channel.Id)
		command := &pb.PeerMessage{
			Channel: channel.Id,
			Peer:    command.Peer,
			Data:    &pb.PeerMessage_Command{Command: command.GetCommand()},
		}
		go forwardRemote(stream, ch, command)
	}
	return nil
}

func execLocal(stream pb.ExecutorService_ExecServer, command *pb.Message) error {
	// create subprocess for a locally-initiated command
	script := command.Data.(*pb.Message_Command).Command
	peer := command.Peer
	cmd := exec.Command("bash", "-c", script)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// handle stdout, and stderr
	handleStream := func(reader io.Reader, transform func([]byte) *pb.Result) {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("Error reading stream: %v", err)
				return
			}
			if err := stream.Send(transform(buf[:n])); err != nil {
				log.Printf("Error sending stream: %v", err)
				return
			}
		}
	}

	go handleStream(stdout, func(data []byte) *pb.Result {
		return &pb.Result{Peer: peer, Data: &pb.Result_Stdout{Stdout: data}}
	})

	go handleStream(stderr, func(data []byte) *pb.Result {
		return &pb.Result{Peer: peer, Data: &pb.Result_Stderr{Stderr: data}}
	})

	go func() {
		for {
			message, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
			chunk := message.Data.(*pb.Message_Stdin).Stdin
			stdin.Write(chunk)
		}
	}()
	return nil
}

func forwardRemote(stream pb.ExecutorService_ExecServer, channel chan *pb.PeerMessage, command *pb.PeerMessage) {
	// ch has the exec command added to it, it will be automatically sent to router
	// take care of stdin, stdout, and stderr
	go func() {
		for {
			message, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
			channel <- &pb.PeerMessage{
				Peer:    command.Peer,
				Channel: command.Channel,
				Data:    &pb.PeerMessage_Stdin{Stdin: message.Data.(*pb.Message_Stdin).Stdin}}
		}
	}()
	for msg := range channel {
		stdout := msg.Data.(*pb.PeerMessage_Stdout).Stdout
		if stdout != nil {
			stream.Send(&pb.Result{
				Peer: msg.Peer,
				Data: &pb.Result_Stdout{Stdout: stdout},
			})
		}
		stderr := msg.Data.(*pb.PeerMessage_Stderr).Stderr
		if stderr != nil {
			stream.Send(&pb.Result{
				Peer: msg.Peer,
				Data: &pb.Result_Stderr{Stderr: stderr},
			})
		}
	}
}

func execRemote(channel chan *pb.PeerMessage, command *pb.PeerMessage) error {
	// create a subprocess for a locally-initiated command
	script := command.Data.(*pb.PeerMessage_Command).Command
	cmd := exec.Command("bash", "-c", script)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// handle stdout, and stderr
	handleStream := func(reader io.Reader, transform func(string, string, []byte) *pb.PeerMessage) {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			channel <- transform(command.Channel, command.Peer, buf[:n])
		}
	}

	go handleStream(stdout, func(id string, peer string, data []byte) *pb.PeerMessage {
		return &pb.PeerMessage{Channel: id, Peer: peer, Data: &pb.PeerMessage_Stdout{Stdout: data}}
	})

	go handleStream(stderr, func(id string, peer string, data []byte) *pb.PeerMessage {
		return &pb.PeerMessage{Channel: id, Peer: peer, Data: &pb.PeerMessage_Stderr{Stderr: data}}
	})

	go func() {
		for message := range channel {
			chunk := message.Data.(*pb.PeerMessage_Stdin).Stdin
			stdin.Write(chunk)
		}
	}()
	return nil
}

func Start() {

	eSig := make(chan struct{})
	rSig := make(chan struct{})

	// Create server to process executor requests
	go func() {
		defer close(eSig)

		s := grpc.NewServer()
		pb.RegisterExecutorServiceServer(s, &executorServer{})

		// Serve gRPC services on agent.sock
		lis, err := net.Listen("unix", "agent.sock")
		if err != nil {
			log.Printf("failed to listen: %v", err)
			return
		}
		defer lis.Close()

		log.Println("Executor service running on unix:agent.sock")
		if s.Serve(lis) != nil {
			log.Printf("failed to serve: %v", err)
			return
		}
		defer s.GracefulStop()
	}()

	// Create client to connect to router
	go func() {
		defer close(rSig)

		routerUrl := "unix:router.sock"
		// Setup gRPC services
		conn, err := grpc.NewClient(routerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		channelSvcClient = pb.NewChannelServiceClient(conn)
		routerSvcClient := pb.NewRouterServiceClient(conn)

		func() {
			stream, err := routerSvcClient.Connect(context.Background())
			if err != nil {
				log.Fatalf("Error creating stream: %v", err)
			}
			defer stream.CloseSend()
			bus = CreateBus(stream)
		}()

		// intercept commands from router
		go func() {
			intercept := bus.Intercept()
			for message := range intercept {
				ch := bus.Channel(message.Channel)
				execRemote(ch, message)
			}
		}()
	}()

	<-eSig
	fmt.Println("Executor service shut down...")
	<-rSig
	fmt.Println("Routing service shut down...")

	fmt.Print("Exiting...")
}
