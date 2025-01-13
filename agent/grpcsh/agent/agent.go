package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"sync"

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
var selfId string
var bufsize = 1024

func (s *executorServer) Exec(stream pb.ExecutorService_ExecServer) error {
	log.Printf("[%s] received Exec command\n", selfId)
	cmd, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive command: %w", err)
	}
	// get command args
	toId := cmd.To
	flag := cmd.Flag
	data := cmd.Data
	// command should always be command
	if flag != pb.Flag_COMMAND {
		return fmt.Errorf("expected command, got: %s", flag.String())
	}
	if toId == selfId {
		// run command locally
		if err := execLocalOnLocal(stream, cmd); err != nil {
			return fmt.Errorf("failed to execute local command: %w", err)
		}
	} else {
		// run command remotely
		ctx := context.Background()
		chnl, err := channelSvcClient.CreateChannel(ctx, &emptypb.Empty{})
		chnlId := chnl.Id
		if err != nil {
			return fmt.Errorf("failed to create channel: %w", err)
		}
		log.Printf("[%s] got channel: %s\n", selfId, chnlId)
		ci, co := bus.Channel(chnlId)
		cmd := &pb.PeerMessage{Channel: chnlId, From: selfId, To: toId, Flag: flag, Data: data}
		err = execLocalOnRemote(stream, ci, co, cmd)
		bus.Close(chnlId)
		if err != nil {
			return fmt.Errorf("failed to forward remote command: %w\n", err)
		}
	}
	return nil
}

func execLocalOnLocal(stream pb.ExecutorService_ExecServer, cmd *pb.Message) error {
	if cmd.Flag != pb.Flag_COMMAND {
		return fmt.Errorf("expected command, got: %s", cmd.Flag.String())
	}
	script := string(cmd.Data)
	proc := exec.Command("bash", "-c", script)
	stdin, err := proc.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	if err := proc.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan bool, 3)
	mu := sync.RWMutex{}

	handleStream := func(reader io.Reader, transform func([]byte, bool) *pb.Result) error {
		buf := make([]byte, bufsize)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("Error reading chunk: %w", err)
				}
				break
			}
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				mu.Lock()
				err := stream.Send(transform(chunk, false))
				mu.Unlock()
				if err != nil {
					return fmt.Errorf("Error sending chunk: %w", err)
				}
			}
		}
		mu.Lock()
		err := stream.Send(transform(nil, true))
		mu.Unlock()
		if err != nil {
			return fmt.Errorf("Error sending EOF: %w", err)
		}
		return nil
	}

	go func() {
		if err := handleStream(stdout, func(data []byte, eof bool) *pb.Result {
			if !eof {
				return &pb.Result{From: selfId, To: selfId, Flag: pb.Flag_MSG_STDOUT, Data: data}
			} else {
				return &pb.Result{From: selfId, To: selfId, Flag: pb.Flag_EOF_STDOUT, Data: nil}
			}
		}); err != nil {
			log.Printf("[%s] error handling stdout stream: %s\n", selfId, err)
		}
		done <- true
	}()

	go func() {
		if err := handleStream(stderr, func(data []byte, eof bool) *pb.Result {
			if !eof {
				return &pb.Result{From: selfId, To: selfId, Flag: pb.Flag_MSG_STDERR, Data: data}
			} else {
				return &pb.Result{From: selfId, To: selfId, Flag: pb.Flag_EOF_STDERR, Data: nil}
			}
		}); err != nil {
			log.Printf("[%s] error handling stderr stream: %s\n", selfId, err)
		}
		done <- true
	}()

	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s] error receiving stream: %s\n", selfId, err)
				}
				break
			}
			if msg.Flag == pb.Flag_EOF_STDIN {
				break
			}
			if msg.Flag != pb.Flag_MSG_STDIN {
				log.Printf("[%s] expected stdin, got: %s\n", selfId, msg.Flag.String())
				break
			}
			chunk := msg.Data
			if n, err := stdin.Write(chunk); err != nil {
				log.Printf("[%s] error writing to stdin: %s\n", selfId, err)
				break
			} else if n != len(chunk) {
				log.Printf("[%s] failed to write all bytes to stdin\n", selfId)
				break
			}
		}
		log.Printf("[%s] done handling stdin stream\n", selfId)
		stdin.Close()
		done <- true
	}()

	// wait for all goroutines to finish
	for i := 0; i < 3; i++ {
		<-done
	}
	proc.Wait()

	log.Printf("[%s] Exec command finished\n", selfId)
	return nil
}

func execLocalOnRemote(stream pb.ExecutorService_ExecServer, in chan *pb.PeerMessage, out chan *pb.PeerMessage, cmd *pb.PeerMessage) error {
	log.Printf("[%s] forwarding remote command: %s\n", selfId, cmd)

	chId := cmd.Channel
	toId := cmd.To

	done := make(chan bool, 2)

	go func() {
		out <- cmd
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s] error receiving stream: %s\n", selfId, err)
				}
				break
			}
			flag := msg.Flag
			data := msg.Data
			if flag == pb.Flag_EOF_STDIN {
				break
			}
			out <- &pb.PeerMessage{Channel: chId, From: selfId, To: toId, Flag: flag, Data: data}
		}
		out <- &pb.PeerMessage{Channel: chId, From: selfId, To: toId, Flag: pb.Flag_EOF_STDIN, Data: nil}
		done <- true
	}()

	go func() {
		for msg := range in {
			flag := msg.Flag
			data := msg.Data
			if flag == pb.Flag_MSG_STDOUT || flag == pb.Flag_MSG_STDERR || flag == pb.Flag_EOF_STDOUT || flag == pb.Flag_EOF_STDERR {
				err := stream.Send(&pb.Result{From: selfId, To: msg.To, Flag: flag, Data: data})
				if err != nil {
					log.Printf("[%s] error sending msg: %s\n", selfId, err)
					break
				}
			} else {
				log.Printf("[%s] unexpected message: %s\n", selfId, msg)
				break
			}
		}
		done <- true
	}()

	for i := 0; i < 2; i++ {
		<-done
	}
	log.Printf("[%s] Exec command finished\n", selfId)
	return nil
}

func execRemoteOnLocal(in chan *pb.PeerMessage, out chan *pb.PeerMessage, cmd *pb.PeerMessage) error {

	flag := cmd.Flag
	if flag != pb.Flag_COMMAND {
		return fmt.Errorf("expected command, got: %s", flag.String())
	}

	chId := cmd.Channel
	to := cmd.From
	script := string(cmd.Data)

	// create a subprocess for a locally-initiated command
	proc := exec.Command("bash", "-c", script)
	log.Printf("[%s] execRemote(): %s\n", selfId, script)
	stdin, err := proc.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := proc.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := proc.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	if err := proc.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan bool, 3)

	handleStream := func(reader io.Reader, transform func(string, string, []byte, bool) *pb.PeerMessage) {
		buf := make([]byte, bufsize)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s] error reading stream: %s\n", selfId, err)
				}
				break
			}
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				out <- transform(chId, to, chunk, false)
			}
		}
		out <- transform(chId, to, nil, true)
		done <- true
	}

	go handleStream(stdout, func(id string, peer string, data []byte, eof bool) *pb.PeerMessage {
		if !eof {
			return &pb.PeerMessage{Channel: id, From: selfId, To: peer, Flag: pb.Flag_MSG_STDOUT, Data: data}
		} else {
			return &pb.PeerMessage{Channel: id, From: selfId, To: peer, Flag: pb.Flag_EOF_STDOUT, Data: nil}
		}
	})

	go handleStream(stderr, func(id string, peer string, data []byte, eof bool) *pb.PeerMessage {
		if !eof {
			return &pb.PeerMessage{Channel: id, From: selfId, To: peer, Flag: pb.Flag_MSG_STDERR, Data: data}
		} else {
			return &pb.PeerMessage{Channel: id, From: selfId, To: peer, Flag: pb.Flag_EOF_STDERR, Data: nil}
		}
	})

	go func() {
		for msg := range in {
			if msg.Flag == pb.Flag_MSG_STDIN {
				stdin.Write(msg.Data)
				continue
			}
			if msg.Flag != pb.Flag_EOF_STDIN {
				log.Printf("[%s] unexpected message: %s\n", selfId, msg)
			}
			break
		}
		log.Printf("[%s] done handling stdin stream\n", selfId)
		stdin.Close()
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}
	proc.Wait()

	log.Printf("[%s] execRemote finished\n", selfId)
	return nil
}

func Start(peerID string, routerUrl string, socketPath string) {

	eSig := make(chan struct{})
	rSig := make(chan struct{})
	selfId = peerID

	// server to process executor requests
	go func() {
		defer close(eSig)

		s := grpc.NewServer()
		pb.RegisterExecutorServiceServer(s, &executorServer{})

		// Serve on socketPath
		lis, err := net.Listen("unix", socketPath)
		if err != nil {
			log.Printf("[%s] failed to listen: %s\n", selfId, err)
			return
		}
		defer lis.Close()
		log.Printf("[%s] started server on unix://%s\n", selfId, socketPath)

		log.Printf("[%s] starting ExecutorService[gRPC]\n", selfId)
		if s.Serve(lis) != nil {
			log.Printf("[%s] failed to start ExecutorService[gRPC]\n", selfId)
		}
		s.GracefulStop()
	}()

	// connector to router
	go func() {
		defer close(rSig)

		conn, err := grpc.NewClient(routerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("[%s] cannot connect to %s [gRPC]: %s\n", selfId, routerUrl, err)
			return
		}
		defer conn.Close()

		channelSvcClient = pb.NewChannelServiceClient(conn)
		routerSvcClient := pb.NewRouterServiceClient(conn)

		ctx := context.Background()
		stream, err := routerSvcClient.Connect(ctx)
		if err != nil {
			log.Printf("[%s] error creating stream: %s\n", selfId, err)
			return
		}
		defer stream.CloseSend()

		func() {
			err := stream.Send(&pb.PeerMessage{From: selfId})
			if err != nil {
				log.Printf("[%s] error sending peer ID: %s\n", selfId, err)
				return
			}
			log.Printf("[%s] sent peer ID: %s\n", selfId, peerID)
		}()

		bus = CreateBus(stream)
		intercept := bus.Intercept()

		for message := range intercept {
			log.Printf("[%s] received remote command: %s\n", selfId, message)
			go func() {
				ci, co := bus.Channel(message.Channel)
				err := execRemoteOnLocal(ci, co, message)
				bus.Close(message.Channel)
				if err != nil {
					log.Printf("[%s] error executing remote command: %s\n", selfId, err)
				}
			}()
		}
	}()

	<-rSig
	log.Printf("[%s] disconnected from RouterService[gRPC]\n", selfId)
	<-eSig
	log.Printf("[%s] shut down ExecutorService[gRPC]\n", selfId)

	log.Printf("[%s] exiting\n", selfId)
}
