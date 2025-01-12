package grpcsh

import (
	"context"
	"fmt"
	pb "grpcsh/pb"
	"io"
	"log"
	"net"
	"os/exec"

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
	handleStream := func(reader io.Reader, createFn func([]byte) *pb.Result) {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			stream.Send(createFn(buf[:n]))
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
	handleStream := func(reader io.Reader, createFn func(string, string, []byte) *pb.PeerMessage) {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			channel <- createFn(command.Channel, command.Peer, buf[:n])
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

func main() {

	// Start UNIX server to process user-sent commands
	// (b) if peer == local, run command locally (as subprocess)
	// (a) else, relay command to router
	go func() {
		// Setup gRPC services
		s := grpc.NewServer()
		pb.RegisterExecutorServiceServer(s, &executorServer{})

		// Serve gRPC services on agent.sock
		lis, err := net.Listen("unix", "agent.sock")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		defer lis.Close()

		log.Println("Executor service running on unix:agent.sock")
		if s.Serve(lis) != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		defer s.GracefulStop()
	}()

	// Start TCP client to process router-sent commands
	// (c) if peer == local, run command locally (as subprocess)
	// (d) if peer != local, raise error (should not happen)
	var routerChannel = make(chan struct{})
	go func() {
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

	<-routerChannel
	fmt.Print("Exiting...")
}
