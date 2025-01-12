package agent

import pb "grpcsh/pb"

// Router handles message routing between virtual channels
type Router struct {
	channels  map[string]chan *pb.PeerMessage
	bus       pb.RouterService_ConnectClient
	intercept chan *pb.PeerMessage
}

// New creates a router and starts reading from the stream
func CreateBus(bus pb.RouterService_ConnectClient) *Router {
	r := &Router{
		channels: make(map[string]chan *pb.PeerMessage),
		bus:      bus,
	}

	// Start receiving messages
	go func() {
		for {
			peerMessage, err := r.bus.Recv()
			if err != nil {
				return
			}
			// whenever peerMessage is a command, push it to intercept
			if ch := r.channels[peerMessage.Channel]; ch != nil {
				if peerMessage.Data.(*pb.PeerMessage_Command) != nil {
					r.intercept <- peerMessage
				} else {
					ch <- peerMessage
				}
			}
		}
	}()
	return r
}

// Channel gets or creates a channel
func (r *Router) Channel(id string) chan *pb.PeerMessage {
	if ch := r.channels[id]; ch != nil {
		return ch
	}
	ch := make(chan *pb.PeerMessage, 10)
	r.channels[id] = ch

	// Start sender for this channel
	go func() {
		for message := range ch {
			r.bus.Send(message)
		}
	}()
	return ch
}

func (r *Router) Intercept() chan *pb.PeerMessage {
	return r.intercept
}

// Close closes a channel
func (r *Router) Close(id string) {
	if ch := r.channels[id]; ch != nil {
		close(ch)
		delete(r.channels, id)
	}
}
