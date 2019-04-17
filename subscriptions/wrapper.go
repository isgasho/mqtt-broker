package subscriptions

import (
	"io"
	"log"

	"github.com/vx-labs/mqtt-protocol/packet"
)

type UserSubscription interface {
	io.Closer
	Send(*packet.Publish) error
	Metadata() Subscription
}

type Session interface {
	Publish(p *packet.Publish) error
}
type Publisher func(*packet.Publish) error

func RemotePublisher(addr string) Publisher {
	return func(*packet.Publish) error {
		log.Printf("INFO: would have send packet to %s", addr)
		return nil
	}
}
func LocalPublisher(session Session) Publisher {
	return session.Publish
}

type userSubscription struct {
	ID        string
	Tenant    string
	SessionID string
	Peer      string
	pb        *Subscription
	publisher Publisher
	closer    func() error
}

func Local(publisher Publisher, message *Subscription) *userSubscription {
	return &userSubscription{
		ID:        message.ID,
		Peer:      message.Peer,
		SessionID: message.SessionID,
		Tenant:    message.Tenant,
		publisher: publisher,
		pb:        message,
	}
}
