package subscriptions

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/weaveworks/mesh"
)

type Store interface {
	ByTopic(tenant string, pattern []byte) (SubscriptionList, error)
	ByID(id string) (*Subscription, error)
	All() (SubscriptionList, error)
	ByPeer(peer uint64) (SubscriptionList, error)
	BySession(id string) (SubscriptionList, error)
	Sessions() ([]string, error)
	Create(subscription *Subscription) error
	Delete(id string) error
}
type memDBStore struct {
	db           *memdb.MemDB
	patternIndex *topicIndexer
	gossip       mesh.Gossip
}
type Router interface {
	NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error)
}

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)
var now = func() int64 {
	return time.Now().UnixNano()
}

func NewMemDBStore(router Router) (*memDBStore, error) {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"subscriptions": &memdb.TableSchema{
				Name: "subscriptions",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:         "id",
						AllowMissing: false,
						Unique:       true,
						Indexer: &memdb.StringFieldIndex{
							Field: "ID",
						},
					},
					"tenant": &memdb.IndexSchema{
						Name:         "tenant",
						AllowMissing: false,
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "Tenant"},
					},
					"session": &memdb.IndexSchema{
						Name:         "session",
						AllowMissing: false,
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "SessionID"},
					},
					"peer": &memdb.IndexSchema{
						Name:         "peer",
						AllowMissing: false,
						Unique:       false,
						Indexer:      &memdb.UintFieldIndex{Field: "Peer"},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	s := &memDBStore{
		db:           db,
		patternIndex: TenantTopicIndexer(),
	}
	gossip, err := router.NewGossip("mqtt-subscriptions", s)
	if err != nil {
		return nil, err
	}
	s.gossip = gossip
	return s, nil
}

type topicIndexer struct {
	root *INode
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: NewINode(),
	}
}
func (t *topicIndexer) Remove(tenant, id string, pattern []byte) error {
	return t.root.Remove(tenant, id, Topic(pattern))
}
func (t *topicIndexer) Lookup(tenant string, pattern []byte) (*SubscriptionList, error) {
	set := t.root.Select(tenant, nil, Topic(pattern)).Filter(func(s *Subscription) bool {
		return s.IsAdded()
	})
	return &set, nil
}

func (s *topicIndexer) Index(subscription *Subscription) error {
	s.root.Insert(
		Topic(subscription.Pattern),
		subscription.Tenant,
		subscription,
	)
	return nil
}
func (m *memDBStore) do(write bool, f func(*memdb.Txn) error) error {
	tx := m.db.Txn(write)
	defer tx.Abort()
	return f(tx)
}
func (m *memDBStore) read(f func(*memdb.Txn) error) error {
	return m.do(false, f)
}
func (m *memDBStore) write(f func(*memdb.Txn) error) error {
	return m.do(true, f)
}

func (m *memDBStore) first(tx *memdb.Txn, index string, value ...interface{}) (*Subscription, error) {
	var ok bool
	var res *Subscription
	data, err := tx.First("subscriptions", index, value...)
	if err != nil {
		return res, err
	}
	res, ok = data.(*Subscription)
	if !ok {
		return res, errors.New("invalid type fetched")
	}
	if res.IsRemoved() {
		return nil, ErrSubscriptionNotFound
	}
	return res, nil
}
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (SubscriptionList, error) {
	var set SubscriptionList
	iterator, err := tx.Get("subscriptions", index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(*Subscription)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		if res.IsAdded() {
			set.Subscriptions = append(set.Subscriptions, res)
		}
	}
}

func (m *memDBStore) All() (SubscriptionList, error) {
	var set SubscriptionList
	var err error
	return set, m.read(func(tx *memdb.Txn) error {
		set, err = m.all(tx, "id")
		if err != nil {
			return err
		}
		return nil
	})
}

func (m *memDBStore) ByID(id string) (*Subscription, error) {
	var res *Subscription
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (SubscriptionList, error) {
	var res SubscriptionList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) BySession(session string) (SubscriptionList, error) {
	var res SubscriptionList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "session", session)
		return
	})
}
func (m *memDBStore) ByPeer(peer uint64) (SubscriptionList, error) {
	var res SubscriptionList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "peer", peer)
		return
	})
}
func (m *memDBStore) ByTopic(tenant string, pattern []byte) (*SubscriptionList, error) {
	return m.patternIndex.Lookup(tenant, pattern)
}
func (m *memDBStore) Sessions() ([]string, error) {
	var res SubscriptionList
	err := m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "session")
		return
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, len(res.Subscriptions))
	for idx := range res.Subscriptions {
		out[idx] = res.Subscriptions[idx].SessionID
	}
	return out, nil
}
func (m *memDBStore) Delete(id string) error {
	session, err := m.ByID(id)
	if err != nil {
		return err
	}
	session.LastDeleted = now()
	defer m.gossip.GossipBroadcast(&SubscriptionList{
		Subscriptions: []*Subscription{session},
	})
	return m.insert(session)
}
func MakeSubscriptionID(session string, pattern []byte) (string, error) {
	hash := sha1.New()
	_, err := hash.Write([]byte(session))
	if err != nil {
		return "", err
	}
	_, err = hash.Write(pattern)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (m *memDBStore) insert(message *Subscription) error {
	if message.IsAdded() {
		err := m.patternIndex.Index(message)
		if err != nil {
			return err
		}
	}
	if message.IsRemoved() {
		err := m.patternIndex.Remove(message.Tenant, message.ID, message.Pattern)
		if err != nil {
			return err
		}
	}
	return m.write(func(tx *memdb.Txn) error {
		err := tx.Insert("subscriptions", message)
		if err != nil {
			return err
		}
		tx.Commit()
		return nil
	})
}
func (m *memDBStore) Create(message *Subscription) error {
	if message.ID == "" {
		log.Printf("WARN: autogenerating ID for subscription on topic %s", string(message.Pattern))
		id, err := MakeSubscriptionID(message.SessionID, message.Pattern)
		if err != nil {
			return err
		}
		message.ID = id
	}
	message.LastUpdated = now()
	defer m.gossip.GossipBroadcast(&SubscriptionList{
		Subscriptions: []*Subscription{message},
	})
	return m.insert(message)
}
