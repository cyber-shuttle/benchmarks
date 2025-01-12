package agent

import (
	pb "grpcsh/pb"
	"log"
)

type Bus struct {
	channels_i map[string]chan *pb.PeerMessage
	channels_o map[string]chan *pb.PeerMessage
	stream     pb.RouterService_ConnectClient
	intercept  chan *pb.PeerMessage
}

func CreateBus(stream pb.RouterService_ConnectClient) *Bus {
	b := &Bus{
		channels_i: make(map[string]chan *pb.PeerMessage),
		channels_o: make(map[string]chan *pb.PeerMessage),
		stream:     stream,
	}

	go func() {
		for {
			peerMessage, err := b.stream.Recv()
			if err != nil {
				return
			}
			if ch := b.channels_i[peerMessage.Channel]; ch != nil {
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

func (b *Bus) Channel(id string) (chan *pb.PeerMessage, chan *pb.PeerMessage) {

	ci := b.channels_i[id]
	co := b.channels_o[id]

	if ci != nil && co != nil {
		return ci, co
	}

	b.channels_i[id] = make(chan *pb.PeerMessage, 10)
	b.channels_o[id] = make(chan *pb.PeerMessage, 10)

	go func() {
		for message := range b.channels_o[id] {
			b.stream.Send(message)
		}
	}()

	return ci, co
}

func (b *Bus) Intercept() chan *pb.PeerMessage {
	return b.intercept
}

func (b *Bus) Close(id string) {
	if ch := b.channels_o[id]; ch != nil {
		close(ch)
		delete(b.channels_o, id)
	}
}
