package agent

import (
	pb "grpcsh/pb"
	"log"
)

type Bus struct {
	channels  map[string]chan *pb.PeerMessage
	stream    pb.RouterService_ConnectClient
	intercept chan *pb.PeerMessage
}

func CreateBus(stream pb.RouterService_ConnectClient) *Bus {
	b := &Bus{
		channels: make(map[string]chan *pb.PeerMessage),
		stream:   stream,
	}

	go func() {
		for {
			peerMessage, err := b.stream.Recv()
			if err != nil {
				return
			}
			if ch := b.channels[peerMessage.Channel]; ch != nil {
				if peerMessage.Data.(*pb.PeerMessage_Command) != nil {
					b.intercept <- peerMessage
				} else {
					ch <- peerMessage
				}
			}
		}
	}()
	log.Println("Bus created")
	return b
}

func (b *Bus) Channel(id string) chan *pb.PeerMessage {
	if ch := b.channels[id]; ch != nil {
		return ch
	}
	ch := make(chan *pb.PeerMessage, 10)
	b.channels[id] = ch

	go func() {
		for message := range ch {
			b.stream.Send(message)
		}
	}()
	return ch
}

func (b *Bus) Intercept() chan *pb.PeerMessage {
	return b.intercept
}

func (b *Bus) Close(id string) {
	if ch := b.channels[id]; ch != nil {
		close(ch)
		delete(b.channels, id)
	}
}
