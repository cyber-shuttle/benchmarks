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
var bus *Bus
var channelSvcClient pb.ChannelServiceClient

func (s *executorServer) Exec(stream pb.ExecutorService_ExecServer) error {
	log.Println("Received Exec command")
	command, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive command: %w", err)
	}
	// command should always be command
	peer := command.Peer
	if peer == "local" {
		go execLocal(stream, command)
	} else {
		ctx := context.Background()
		channel, err := channelSvcClient.CreateChannel(ctx, &emptypb.Empty{})
		if err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}
		log.Println("Got channel:", channel.Id)
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

	handleStream := func(reader io.Reader, transform func([]byte) *pb.Result) {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("Error reading stream:", err)
				return
			}
			if err := stream.Send(transform(buf[:n])); err != nil {
				log.Println("Error sending stream:", err)
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
				log.Println("Error receiving stream:", err)
				return
			}
			chunk := message.Data.(*pb.Message_Stdin).Stdin
			stdin.Write(chunk)
		}
	}()
	return nil
}

func forwardRemote(stream pb.ExecutorService_ExecServer, channel chan *pb.PeerMessage, command *pb.PeerMessage) {
	channel <- command
	go func() {
		for {
			message, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("Error receiving stream:", err)
				return
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
			if err := stream.Send(&pb.Result{
				Peer: msg.Peer,
				Data: &pb.Result_Stdout{Stdout: stdout},
			}); err != nil {
				log.Println("Error sending stdout:", err)
				return
			}
		}
		stderr := msg.Data.(*pb.PeerMessage_Stderr).Stderr
		if stderr != nil {
			if err := stream.Send(&pb.Result{
				Peer: msg.Peer,
				Data: &pb.Result_Stderr{Stderr: stderr},
			}); err != nil {
				log.Println("Error sending stderr:", err)
				return
			}
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
				log.Println("Error reading stream:", err)
				return
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

func Start(peerID string, routerUrl string, socketPath string) {

	eSig := make(chan struct{})
	rSig := make(chan struct{})

	// server to process executor requests
	go func() {
		defer close(eSig)

		s := grpc.NewServer()
		pb.RegisterExecutorServiceServer(s, &executorServer{})

		// Serve on socketPath
		lis, err := net.Listen("unix", socketPath)
		if err != nil {
			log.Println("failed to listen:", err)
			return
		}
		defer lis.Close()
		log.Println("started server on unix://" + socketPath)

		log.Println("[gRPC] serving ExecutorService")
		if s.Serve(lis) != nil {
			log.Println("[gRPC] failed to serve gRPC over socket")
		}
		s.GracefulStop()
	}()

	// connector to router
	go func() {
		defer close(rSig)

		conn, err := grpc.NewClient(routerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println("did not connect:", err)
			return
		}
		defer conn.Close()

		channelSvcClient = pb.NewChannelServiceClient(conn)
		routerSvcClient := pb.NewRouterServiceClient(conn)

		ctx := context.Background()
		stream, err := routerSvcClient.Connect(ctx)
		if err != nil {
			log.Println("Error creating stream:", err)
			return
		}
		defer stream.CloseSend()

		func() {
			err := stream.Send(&pb.PeerMessage{Peer: peerID})
			if err != nil {
				log.Println("Error sending peer ID:", err)
				return
			}
			log.Println("Sent peer ID:", peerID)
		}()

		bus = CreateBus(stream)
		intercept := bus.Intercept()

		for message := range intercept {
			ch := bus.Channel(message.Channel)
			go func() {
				if err := execRemote(ch, message); err != nil {
					log.Println("Error executing remote command:", err)
				}
			}()
		}
	}()

	<-rSig
	log.Println("Routing service shut down...")
	<-eSig
	log.Println("Executor service shut down...")

	log.Print("Exiting...")
}
