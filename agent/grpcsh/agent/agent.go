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
var selfPeerId string

func (s *executorServer) Exec(stream pb.ExecutorService_ExecServer) error {
	log.Println("Received Exec command")
	command, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive command: %w", err)
	}
	// command should always be command
	to := command.To
	if to == selfPeerId {
		if err := execLocal(stream, command); err != nil {
			return fmt.Errorf("failed to execute local command: %w", err)
		}
	} else {
		ctx := context.Background()
		channel, err := channelSvcClient.CreateChannel(ctx, &emptypb.Empty{})
		if err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}
		log.Println("Got channel:", channel.Id)
		ci, co := bus.Channel(channel.Id)
		command := &pb.PeerMessage{
			Channel: channel.Id,
			From:    selfPeerId,
			To:      to,
			Data:    &pb.PeerMessage_Command{Command: command.GetCommand()},
		}
		if err := forwardRemote(stream, ci, co, command); err != nil {
			return fmt.Errorf("failed to forward remote command: %w", err)
		}
	}
	return nil
}

func execLocal(stream pb.ExecutorService_ExecServer, command *pb.Message) error {
	script := command.Data.(*pb.Message_Command).Command
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

	done := make(chan bool, 3)

	handleStream := func(reader io.Reader, transform func([]byte) *pb.Result) error {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("Error reading stream: %w", err)
				}
				break
			}
			if err := stream.Send(transform(buf[:n])); err != nil {
				return fmt.Errorf("Error sending stream: %w", err)
			}
		}
		return nil
	}

	go func() {
		if err := handleStream(stdout, func(data []byte) *pb.Result {
			return &pb.Result{From: selfPeerId, To: selfPeerId, Data: &pb.Result_Stdout{Stdout: data}}
		}); err != nil {
			log.Println("Error handling stdout stream:", err)
		}
		done <- true
	}()

	go func() {
		if err := handleStream(stderr, func(data []byte) *pb.Result {
			return &pb.Result{From: selfPeerId, To: selfPeerId, Data: &pb.Result_Stderr{Stderr: data}}
		}); err != nil {
			log.Println("Error handling stderr stream:", err)
		}
		done <- true
	}()

	go func() {
		for {
			message, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Println("Error receiving stream:", err)
				}
				break
			}
			chunk := message.Data.(*pb.Message_Stdin).Stdin
			if n, err := stdin.Write(chunk); err != nil {
				log.Println("Error writing to stdin:", err)
				break
			} else if n != len(chunk) {
				log.Println("Failed to write all bytes to stdin")
				break
			}
		}
		log.Println("Done handling stdin stream")
		done <- true
	}()

	// wait for all goroutines to finish
	cmd.Wait()
	for i := 0; i < 3; i++ {
		<-done
	}
	log.Println("Exec command finished")
	return nil
}

func forwardRemote(stream pb.ExecutorService_ExecServer, in chan *pb.PeerMessage, out chan *pb.PeerMessage, command *pb.PeerMessage) error {
	fmt.Println("Forwarding remote command", command)
	done := make(chan bool, 1)

	go func() {
		out <- command
		for {
			message, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Println("Error receiving stream:", err)
				}
				break
			}
			out <- &pb.PeerMessage{
				From:    selfPeerId,
				To:      command.To,
				Channel: command.Channel,
				Data:    &pb.PeerMessage_Stdin{Stdin: message.Data.(*pb.Message_Stdin).Stdin}}
		}
		done <- true
	}()

	for msg := range in {

		if data, ok := msg.Data.(*pb.PeerMessage_Stdout); ok {
			if err := stream.Send(&pb.Result{
				From: selfPeerId,
				To:   msg.To,
				Data: &pb.Result_Stdout{Stdout: data.Stdout},
			}); err != nil {
				log.Println("Error sending stdout:", err)
				return err
			}
		}
		if data, ok := msg.Data.(*pb.PeerMessage_Stderr); ok {
			if err := stream.Send(&pb.Result{
				From: selfPeerId,
				To:   msg.To,
				Data: &pb.Result_Stderr{Stderr: data.Stderr},
			}); err != nil {
				log.Println("Error sending stderr:", err)
				return err
			}
		}
	}
	<-done
	log.Println("Exec command finished")
	return nil
}

func execRemote(in chan *pb.PeerMessage, out chan *pb.PeerMessage, command *pb.PeerMessage) error {
	// create a subprocess for a locally-initiated command
	script := command.Data.(*pb.PeerMessage_Command).Command
	to := command.From
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
				if err != io.EOF {
					log.Println("Error reading stream:", err)
				}
				break
			}
			out <- transform(command.Channel, to, buf[:n])
		}
	}

	go handleStream(stdout, func(id string, peer string, data []byte) *pb.PeerMessage {
		return &pb.PeerMessage{Channel: id, From: selfPeerId, To: peer, Data: &pb.PeerMessage_Stdout{Stdout: data}}
	})

	go handleStream(stderr, func(id string, peer string, data []byte) *pb.PeerMessage {
		return &pb.PeerMessage{Channel: id, From: selfPeerId, To: peer, Data: &pb.PeerMessage_Stderr{Stderr: data}}
	})

	go func() {
		for message := range in {
			chunk := message.Data.(*pb.PeerMessage_Stdin).Stdin
			stdin.Write(chunk)
		}
	}()
	cmd.Wait()
	log.Println("Command finished")
	return nil
}

func Start(peerID string, routerUrl string, socketPath string) {

	eSig := make(chan struct{})
	rSig := make(chan struct{})
	selfPeerId = peerID

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
			err := stream.Send(&pb.PeerMessage{From: selfPeerId})
			if err != nil {
				log.Println("Error sending peer ID:", err)
				return
			}
			log.Println("Sent peer ID:", peerID)
		}()

		bus = CreateBus(stream)
		intercept := bus.Intercept()

		for message := range intercept {
			fmt.Println("Intercepted message", message)
			ci, co := bus.Channel(message.Channel)
			go func() {
				if err := execRemote(ci, co, message); err != nil {
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
