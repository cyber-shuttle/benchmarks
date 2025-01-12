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
		intercept:  make(chan *pb.PeerMessage),
	}

	go func() {
		for {
			peerMessage, err := b.stream.Recv()
			log.Printf("[%s] mux received: %s\n", selfPeerId, peerMessage)
			if err != nil {
				return
			}
			if _, ok := peerMessage.Data.(*pb.PeerMessage_Command); ok {
				b.intercept <- peerMessage
			} else {
				c := peerMessage.Channel
				if ch, exists := b.channels_i[c]; exists {
					ch <- peerMessage
				}
			}
		}
	}()
	log.Printf("[%s] mux created bus\n", selfPeerId)
	return b
}

func (b *Bus) Channel(id string) (chan *pb.PeerMessage, chan *pb.PeerMessage) {
	log.Printf("[%s] mux received bi-channel request: %s\n", selfPeerId, id)

	ci := b.channels_i[id]
	co := b.channels_o[id]

	if ci != nil && co != nil {
		log.Printf("[%s] mux returning existing bi-channels: %s\n", selfPeerId, id)
		return ci, co
	}

	log.Printf("[%s] mux returning new bi-channels: %s\n", selfPeerId, id)
	ci = make(chan *pb.PeerMessage)
	co = make(chan *pb.PeerMessage)
	b.channels_i[id] = ci
	b.channels_o[id] = co

	go func() {
		for message := range co {
			log.Printf("[%s] mux sending: %s\n", selfPeerId, message)
			if err := b.stream.Send(message); err != nil {
				log.Printf("[%s] mux got error when sending: %s\n", selfPeerId, err)
				return
			}
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
