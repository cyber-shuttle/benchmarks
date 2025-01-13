package agent

import (
	pb "grpcsh/pb"
	"log"
	"sync"
)

type Bus struct {
	channels_i map[string]chan *pb.PeerMessage
	channels_o map[string]chan *pb.PeerMessage
	stream     pb.RouterService_ConnectClient
	intercept  chan *pb.PeerMessage
	mu         sync.RWMutex
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
			if peerMessage.Flag == pb.Flag_COMMAND {
				b.intercept <- peerMessage
			} else {
				c := peerMessage.Channel
				b.mu.Lock()
				ch, exists := b.channels_i[c]
				if !exists {
					ch = make(chan *pb.PeerMessage)
					b.channels_i[c] = ch
				}
				b.mu.Unlock()
				ch <- peerMessage
			}
		}
	}()
	log.Printf("[%s] mux created bus\n", selfPeerId)
	return b
}

func (b *Bus) Channel(id string) (chan *pb.PeerMessage, chan *pb.PeerMessage) {
	log.Printf("[%s] mux received bi-channel request: %s\n", selfPeerId, id)

	begin_channel_loop := false

	b.mu.Lock()

	if b.channels_i[id] == nil {
		b.channels_i[id] = make(chan *pb.PeerMessage)
	}

	if b.channels_o[id] == nil {
		b.channels_o[id] = make(chan *pb.PeerMessage)
		begin_channel_loop = true
	}

	b.mu.Unlock()

	if begin_channel_loop {
		go func() {
			for message := range b.channels_o[id] {
				log.Printf("[%s] mux sending: %s\n", selfPeerId, message)
				if err := b.stream.Send(message); err != nil {
					log.Printf("[%s] mux got error when sending: %s\n", selfPeerId, err)
					return
				}
			}
		}()
	}
	return b.channels_i[id], b.channels_o[id]
}

func (b *Bus) Intercept() chan *pb.PeerMessage {
	return b.intercept
}

func (b *Bus) Close(id string) {
	if ch, exists := b.channels_o[id]; exists {
		close(ch)
		delete(b.channels_o, id)
	}
	if ch, exists := b.channels_i[id]; exists {
		close(ch)
		delete(b.channels_i, id)
	}
}
