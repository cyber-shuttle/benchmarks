package router

import (
	"context"
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
	log.Println("Received connect request")
	// received peer ID
	peer, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive peer ID: %w", err)
	}
	peerId := peer.Peer
	log.Println("Received peer ID:", peerId)

	// Assign peer ID
	s.mu.Lock() // Write lock when modifying the map and counter
	s.peers[peerId] = stream
	s.mu.Unlock()

	log.Println("Connected peer:", peerId)
	defer func() {
		s.mu.Lock() // Write lock when removing from map
		delete(s.peers, peerId)
		s.mu.Unlock()
		log.Println("Disconnected peer:", peerId)
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			return nil
		}

		if msg.Peer != "" {
			s.mu.RLock() // Read lock when accessing map for routing
			if peer := s.peers[msg.Peer]; peer != nil {
				peer.Send(msg)
			}
			s.mu.RUnlock()
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
	channelId := fmt.Sprintf("channel-%d", len(c.channels)+1)
	c.channels[channelId] = channelId
	c.mu.Unlock()
	log.Println("Created channel:", channelId)
	return &pb.Channel{Id: channelId}, nil
}

func (c *ChannelService) DeleteChannel(ctx context.Context, req *pb.Channel) (*emptypb.Empty, error) {
	c.mu.Lock()
	delete(c.channels, req.Id)
	c.mu.Unlock()
	log.Println("Deleted channel:", req.Id)
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
	log.Println("Server started on:", routerUrl)
	server.Serve(lis)
}
