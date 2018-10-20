package broker

import (
	"io"
	"log"
	"strings"
	"sync"

	"github.com/vx-labs/mqtt-broker/sessions"
	"github.com/vx-labs/mqtt-broker/topics"

	"github.com/vx-labs/mqtt-broker/broker/listener/transport"

	"github.com/weaveworks/mesh"

	"github.com/vx-labs/mqtt-protocol/packet"

	"github.com/vx-labs/mqtt-broker/broker/listener"

	"github.com/golang/protobuf/proto"

	"github.com/vx-labs/mqtt-broker/identity"

	"github.com/vx-labs/mqtt-broker/broker/peer"
	"github.com/vx-labs/mqtt-broker/subscriptions"
)

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/broker/ --go_out=plugins=grpc:. events.proto

type SessionStore interface {
	ById(id string) (*sessions.Session, error)
	ByPeer(peer uint64) (sessions.SessionList, error)
	All() (sessions.SessionList, error)
	Exists(id string) bool
	Upsert(sess *sessions.Session) error
	Delete(id string) error
}

type TopicStore interface {
	Create(message *topics.RetainedMessage) error
	ByTopicPattern(tenant string, pattern []byte) (topics.RetainedMessageList, error)
	All() (topics.RetainedMessageList, error)
}
type SubscriptionStore interface {
	ByTopic(tenant string, pattern []byte) (subscriptions.SubscriptionList, error)
	ByID(id string) (*subscriptions.Subscription, error)
	All() (subscriptions.SubscriptionList, error)
	ByPeer(peer uint64) (subscriptions.SubscriptionList, error)
	BySession(id string) (subscriptions.SubscriptionList, error)
	Sessions() ([]string, error)
	Create(subscription *subscriptions.Subscription) error
	Delete(id string) error
}
type Broker struct {
	authHelper    func(transport listener.Transport, sessionID, username string, password string) (tenant string, id string, err error)
	Peer          *peer.Peer
	Subscriptions SubscriptionStore
	Sessions      SessionStore
	Topics        TopicStore
	localSessions map[string]*listener.Session
	mutex         sync.RWMutex
	Listener      io.Closer
	TCPTransport  io.Closer
	TLSTransport  io.Closer
	WSSTransport  io.Closer
}

func New(id identity.Identity, config Config) *Broker {
	subscriptionsStore, err := subscriptions.NewMemDBStore()
	if err != nil {
		log.Fatal(err)
	}
	topicssStore, err := topics.NewMemDBStore()
	if err != nil {
		log.Fatal(err)
	}
	broker := &Broker{
		Subscriptions: subscriptionsStore,
		Topics:        topicssStore,
		localSessions: map[string]*listener.Session{},
		Sessions:      sessions.NewSessionStore(),
		authHelper:    config.AuthHelper,
	}
	broker.Peer = peer.NewPeer(id, broker.onAdd, broker.onDel, broker.onPeerDown, broker.onUnicast)
	l, listenerCh := listener.New(broker)
	if config.TCPPort > 0 {
		tcpTransport, err := transport.NewTCPTransport(config.TCPPort, listenerCh)
		broker.TCPTransport = tcpTransport
		if err != nil {
			log.Printf("WARN: failed to start TCP listener on port %d: %v", config.TCPPort, err)
		} else {
			log.Printf("INFO: started TCP listener on port %d", config.TCPPort)
		}
	}
	if config.TLS != nil {
		if config.WSSPort > 0 {
			wssTransport, err := transport.NewWSSTransport(config.WSSPort, config.TLS, listenerCh)
			broker.WSSTransport = wssTransport
			if err != nil {
				log.Printf("WARN: failed to start WSS listener on port %d: %v", config.TLSPort, err)
			} else {
				log.Printf("INFO: started WSS listener on port %d", config.TLSPort)
			}
		}
		if config.TLSPort > 0 {
			tlsTransport, err := transport.NewTLSTransport(config.TLSPort, config.TLS, listenerCh)
			broker.TLSTransport = tlsTransport
			if err != nil {
				log.Printf("WARN: failed to start TLS listener on port %d: %v", config.TLSPort, err)
			} else {
				log.Printf("INFO: started TLS listener on port %d", config.TLSPort)
			}
		} else {
			log.Printf("WARN: failed to start TLS listener: TLS config not found")
		}
	}
	broker.Listener = l
	return broker
}
func (b *Broker) decodeEvent(payload string) (*StateEvent, error) {
	event := &StateEvent{}
	return event, proto.Unmarshal([]byte(payload), event)
}
func (b *Broker) onUnicast(payload []byte) {
	message := &MessagePublished{}
	err := proto.Unmarshal(payload, message)
	if err != nil {
		return
	}
	b.dispatch(message)
}
func (b *Broker) onAdd(payload string) {
	event, err := b.decodeEvent(payload)
	if err != nil {
		log.Printf("ERR: failed to decode added event: %v", err)
		return
	}
	switch event.Name {
	case "sessions":
		sess, err := b.Sessions.ById(event.GetSession().ID)
		if err == nil && sess.Peer != uint64(b.Peer.Name()) {
			b.closeLocalSession(sess)
		}
		b.Sessions.Upsert(event.GetSession())
	case "subscriptions":
		b.Subscriptions.Create(event.GetSubscription())
	case "topics":
		b.Topics.Create(event.GetRetainedMessage())
	default:
		log.Printf("WARN: received unhandled event %s", event.Name)
	}
}
func encodeEvent(ev *StateEvent) string {
	payload, err := proto.Marshal(ev)
	if err != nil {
		panic(err)
	}
	return string(payload)
}
func (b *Broker) onPeerDown(name mesh.PeerName) {
	log.Printf("INFO: lost peer %s", name.String())
	members := []string{}
	for _, peer := range b.Peer.Members() {
		if peer > 0 {
			members = append(members, peer.String())
		}
	}
	log.Printf("INFO: current mesh members: %s", strings.Join(members, ","))

	set, err := b.Subscriptions.ByPeer(uint64(name))
	if err != nil {
		log.Printf("ERR: failed to remove subscriptions from peer %d: %v", name, err)
		return
	}
	set.Apply(func(sub *subscriptions.Subscription) {
		b.Subscriptions.Delete(sub.ID)
		b.Peer.Del(encodeEvent(&StateEvent{
			Name:         "subscriptions",
			Subscription: sub,
		}))
	})

	sessionSet, err := b.Sessions.ByPeer(uint64(name))
	if err != nil {
		log.Printf("ERR: failed to fetch sessions from peer %d: %v", name, err)
		return
	}
	sessionSet.Apply(func(s *sessions.Session) {
		if s.WillRetain {
			retainedMessage := &topics.RetainedMessage{
				Payload: s.WillPayload,
				Qos:     s.WillQoS,
				Tenant:  s.Tenant,
				Topic:   s.WillTopic,
			}
			b.Topics.Create(retainedMessage)
			b.Peer.Add(encodeEvent(&StateEvent{
				Name:            "topics",
				RetainedMessage: retainedMessage,
			}))
		}
		recipients, err := b.Subscriptions.ByTopic(s.Tenant, s.WillTopic)
		if err != nil {
			return
		}

		message := &MessagePublished{
			Payload:   s.WillPayload,
			Topic:     s.WillTopic,
			Qos:       make([]int32, 0, len(recipients)),
			Recipient: make([]string, 0, len(recipients)),
		}
		recipients.Apply(func(sub *subscriptions.Subscription) {
			message.Recipient = append(message.Recipient, sub.SessionID)
			qos := s.WillQoS
			if qos > sub.Qos {
				qos = sub.Qos
			}
			message.Qos = append(message.Qos, qos)
		})
		b.dispatch(message)
		b.Sessions.Delete(s.ID)
		b.Peer.Add(encodeEvent(&StateEvent{
			Name:    "sessions",
			Session: s,
		}))
	})
}
func (b *Broker) onDel(payload string) {
	event, err := b.decodeEvent(payload)
	if err != nil {
		log.Printf("ERR: failed to decode deleted event: %v", err)
		return
	}
	switch event.Name {
	case "sessions":
		b.Sessions.Delete(event.GetSession().ID)
	case "topics":
	case "subscriptions":
		b.Subscriptions.Delete(event.GetSubscription().ID)
	default:
		log.Printf("WARN: received unhandled event %s", event.Name)
	}
}
func (b *Broker) Join(hosts []string) {
	log.Printf("INFO: joining hosts %v", hosts)
	b.Peer.Join(hosts)
}

func (b *Broker) dispatch(message *MessagePublished) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for idx, recipient := range message.Recipient {
		packet := &packet.Publish{
			Header: &packet.Header{
				Dup:    message.Dup,
				Qos:    message.Qos[idx],
				Retain: message.Retained,
			},
			Payload:   message.Payload,
			Topic:     message.Topic,
			MessageId: 1,
		}
		if ch, ok := b.localSessions[recipient]; ok {
			select {
			case ch.Channel() <- packet:
			default:
			}
		}
	}
}

func (b *Broker) Stop() {
	log.Printf("INFO: stopping Listener aggregator")
	b.Listener.Close()
	log.Printf("INFO: stopping Listener aggregator stopped")
	if b.TCPTransport != nil {
		log.Printf("INFO: stopping TCP listener")
		b.TCPTransport.Close()
		log.Printf("INFO: TCP listener stopped")
	}
	if b.TLSTransport != nil {
		log.Printf("INFO: stopping TLS listener")
		b.TLSTransport.Close()
		log.Printf("INFO: TLS listener stopped")
	}
	if b.WSSTransport != nil {
		log.Printf("INFO: stopping WSS listener")
		b.WSSTransport.Close()
		log.Printf("INFO: WSS listener stopped")
	}
	b.mutex.Lock()
	if len(b.localSessions) > 0 {
		log.Printf("INFO: Closing client connections")
		for _, session := range b.localSessions {
			session.Close()
		}
		log.Printf("INFO: client connections closed")
	}
	b.mutex.Unlock()

}
