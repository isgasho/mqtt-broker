package broker

import (
	"io"
	"log"

	"github.com/google/uuid"

	"github.com/vx-labs/mqtt-broker/broker/rpc"

	"github.com/vx-labs/mqtt-protocol/packet"
)

type Subscription interface {
	io.Closer
	Send(*packet.Publish) error
}

type Session interface {
	Publish(p *packet.Publish) error
}
type Publisher func(*packet.Publish) error

func RemotePublisher(addr string, caller *rpc.Caller) Publisher {
	return func(*packet.Publish) error {
		log.Printf("INFO: would have send packet to %s", addr)
		return nil
	}
}
func LocalPublisher(session Session) Publisher {
	return session.Publish
}

type subscription struct {
	id        string
	publisher Publisher
	closer    func() error
}

func NewSubscription() *subscription {
	return &subscription{
		id: uuid.New().String(),
	}
}
