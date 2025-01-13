package router

import (
	"context"
	"crypto/md5"
	"fmt"
	pb "grpcsh/pb"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RouterService struct {
	pb.UnimplementedRouterServiceServer
	peers map[string]pb.RouterService_ConnectServer
	mu    sync.RWMutex
}

func (s *RouterService) Connect(stream pb.RouterService_ConnectServer) error {
	log.Printf("[Router] received connect request\n")
	// received peer ID
	peer, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to get peerId: %w", err)
	}
	peerId := peer.From
	log.Printf("[Router] got peerId: %s\n", peerId)

	// Assign peer ID
	s.mu.Lock() // Write lock when modifying the map and counter
	s.peers[peerId] = stream
	s.mu.Unlock()

	log.Printf("[Router] saved peerId: %s\n", peerId)
	defer func() {
		s.mu.Lock() // Write lock when removing from map
		delete(s.peers, peerId)
		s.mu.Unlock()
		log.Printf("[Router] disconnected peerId: %s\n", peerId)
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}
		from := msg.From
		to := msg.To
		if to != "" {
			s.mu.RLock()
			if peer, exists := s.peers[msg.To]; exists {
				log.Printf("[Router] %s -> %s: channel=%s, flag=%s, hash=%x, length=%d\n", from, to, msg.Channel, msg.Flag, md5.Sum(msg.Data), len(msg.Data))
				if err := peer.Send(msg); err != nil {
					log.Printf("[Router] failed to send message: %s\n", err)
				}
			}
			s.mu.RUnlock()
		} else {
			log.Printf("[Router] %s -> [no recipient]: %s\n", from, msg)
		}
	}
}

type ChannelService struct {
	pb.UnimplementedChannelServiceServer
	channels map[string]string
	mu       sync.RWMutex
}

func (c *ChannelService) CreateChannel(ctx context.Context, req *emptypb.Empty) (*pb.Channel, error) {
	c.mu.Lock()
	channelId := fmt.Sprintf("ch%d", len(c.channels)+1)
	c.channels[channelId] = channelId
	c.mu.Unlock()
	log.Printf("[Router] created channel: %s\n", channelId)
	return &pb.Channel{Id: channelId}, nil
}

func (c *ChannelService) DeleteChannel(ctx context.Context, req *pb.Channel) (*emptypb.Empty, error) {
	c.mu.Lock()
	delete(c.channels, req.Id)
	c.mu.Unlock()
	log.Printf("[Router] deleted channel: %s\n", req.Id)
	return &emptypb.Empty{}, nil
}

func Start(routerUrl string) {
	server := grpc.NewServer()
	pb.RegisterRouterServiceServer(server, &RouterService{
		peers: make(map[string]pb.RouterService_ConnectServer),
	})
	pb.RegisterChannelServiceServer(server, &ChannelService{
		channels: make(map[string]string),
	})

	lis, _ := net.Listen("tcp", routerUrl)
	log.Printf("[Router] started on: %s\n", routerUrl)
	server.Serve(lis)
}
