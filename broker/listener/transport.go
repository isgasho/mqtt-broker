package listener

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/vx-labs/mqtt-protocol/decoder"
	"github.com/vx-labs/mqtt-protocol/encoder"
	"github.com/vx-labs/mqtt-protocol/packet"
)

type TimeoutReadWriteCloser interface {
	SetDeadline(time.Time) error
	io.ReadWriteCloser
}
type Transport interface {
	io.Closer
	Name() string
	Encrypted() bool
	EncryptionState() *tls.ConnectionState
	Channel() TimeoutReadWriteCloser
	RemoteAddress() string
}
type Handler interface {
	Authenticate(transport Transport, sessionID []byte, username string, password string) (tenant string, id string, err error)
	OnConnect(sess *Session) (int32, error)
}

type listener struct {
	ch      chan Transport
	handler Handler
}

func (l *listener) Close() error {
	close(l.ch)
	return nil
}

func New(handler Handler, inflightSize int) (io.Closer, chan<- Transport) {
	ch := make(chan Transport)

	l := &listener{
		ch:      ch,
		handler: handler,
	}
	go func() {
		for transport := range ch {
			go l.runSession(transport, inflightSize)
		}
	}()
	return l, l.ch
}

func validateClientID(clientID []byte) bool {
	return len(clientID) > 0 && len(clientID) < 128
}

func (l *listener) runSession(t Transport, inflightSize int) {
	log.Printf("INFO: listener %s: accepted new connection from %s", t.Name(), t.RemoteAddress())
	c := t.Channel()
	handler := l.handler
	session := newSession(t, inflightSize)
	session.encoder = encoder.New(c)
	defer close(session.quit)
	dec := decoder.New(
		decoder.OnConnect(func(p *packet.Connect) error {
			session.keepalive = p.KeepaliveTimer
			session.RenewDeadline()
			clientID := p.ClientId
			if !validateClientID(clientID) {
				session.ConnAck(packet.CONNACK_REFUSED_IDENTIFIER_REJECTED)
				return fmt.Errorf("invalid client ID")
			}
			tenant, id, err := handler.Authenticate(t, clientID, string(p.Username), string(p.Password))
			if err != nil {
				session.encoder.ConnAck(&packet.ConnAck{
					Header:     p.Header,
					ReturnCode: packet.CONNACK_REFUSED_BAD_USERNAME_OR_PASSWORD,
				})
				session.ConnAck(packet.CONNACK_REFUSED_BAD_USERNAME_OR_PASSWORD)
				return fmt.Errorf("authentication failed")
			}
			session.id = id
			session.connect = p
			session.tenant = tenant
			code, err := handler.OnConnect(session)
			if err == nil {
				session.RenewDeadline()
			}
			return session.ConnAck(code)
		}),
		decoder.OnPublish(func(p *packet.Publish) error {
			session.RenewDeadline()
			err := session.incoming.Insert(p)
			if err != nil {
				log.Printf("WARN: failed to accept incomming message from session %s: %v", session.id, err)
				return nil
			}
			if p.Header.Qos == 1 {
				return session.PubAck(p.MessageId)
			}
			return nil
		}),
		decoder.OnSubscribe(func(p *packet.Subscribe) error {
			session.RenewDeadline()
			session.emitSubscribe(p)
			return nil
		}),
		decoder.OnUnsubscribe(func(p *packet.Unsubscribe) error {
			session.RenewDeadline()
			session.emitUnsubscribe(p)
			return nil
		}),
		decoder.OnPubAck(func(p *packet.PubAck) error {
			session.RenewDeadline()
			session.queue.Acknowledge(p.MessageId)
			return nil
		}),
		decoder.OnPingReq(func(p *packet.PingReq) error {
			session.RenewDeadline()
			return session.encoder.PingResp(&packet.PingResp{
				Header: p.Header,
			})
		}),
		decoder.OnDisconnect(func(p *packet.Disconnect) error {
			session.RenewDeadline()
			return ErrSessionDisconnected
		}),
	)
	c.SetDeadline(
		time.Now().Add(15 * time.Second),
	)
	decoderCh := make(chan struct{})
	var err error
	go func() {
		defer close(decoderCh)
		for {
			err = dec.Decode(c)
			if err != nil {
				if err == io.EOF || err == ErrSessionDisconnected || session.closed {
					return
				}
				if opErr, ok := err.(*net.OpError); ok {
					if opErr.Timeout() {
						log.Printf("ERR: listener %s: session_id=%q timedout", t.Name(), session.id)
						return
					}
				}
				log.Printf("ERR: listener %s: decoding from session_id=%q failed: %v", t.Name(), session.id, err)
				return
			}
		}
	}()

	select {
	case <-decoderCh:
		c.Close()
	}
	if err != nil && err != ErrSessionDisconnected && !session.closed {
		log.Printf("WARN: listener %s: session %s lost: %v", t.Name(), session.id, err)
		session.emitLost()
	} else {
		log.Printf("INFO: listener %s: session %s closed", t.Name(), session.id)
		session.emitClosed()
	}
}
