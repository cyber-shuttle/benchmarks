package agent

import (
	"crypto/md5"
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
			msg, err := b.stream.Recv()
			log.Printf("[%s] mux received: %s<-%s, channel=%s, flag=%s, hash=%x, length=%d\n", selfId, msg.To, msg.From, msg.Channel, msg.Flag.String(), md5.Sum(msg.Data), len(msg.Data))
			if err != nil {
				return
			}
			if msg.Flag == pb.Flag_COMMAND {
				b.intercept <- msg
			} else {
				c := msg.Channel
				b.mu.Lock()
				ch, exists := b.channels_i[c]
				if !exists {
					ch = make(chan *pb.PeerMessage)
					b.channels_i[c] = ch
				}
				b.mu.Unlock()
				ch <- msg
			}
		}
	}()
	log.Printf("[%s] mux created bus\n", selfId)
	return b
}

func (b *Bus) Channel(id string) (chan *pb.PeerMessage, chan *pb.PeerMessage) {
	log.Printf("[%s] mux received bi-channel request: %s\n", selfId, id)

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
			ch := b.channels_o[id]
			for msg := range ch {
				log.Printf("[%s] mux sending: %s->%s, channel=%s, flag=%s, hash=%x, length=%d\n", selfId, msg.From, msg.To, msg.Channel, msg.Flag.String(), md5.Sum(msg.Data), len(msg.Data))
				if err := b.stream.Send(msg); err != nil {
					log.Printf("[%s] mux got error when sending: %s\n", selfId, err)
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
