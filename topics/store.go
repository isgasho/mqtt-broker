package topics

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-memdb"
	"github.com/vx-labs/mqtt-broker/state"
	"github.com/weaveworks/mesh"
)

type memDBStore struct {
	db         *memdb.MemDB
	state      *state.Store
	topicIndex *topicIndexer
}

var (
	ErrRetainedMessageNotFound = errors.New("retained message not found")
)

func MakeTopicID(tenant string, topic []byte) (string, error) {
	hash := sha1.New()
	_, err := hash.Write([]byte(tenant))
	if err != nil {
		return "", err
	}
	_, err = hash.Write(topic)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

var now = func() int64 {
	return time.Now().UnixNano()
}

type Router interface {
	NewGossip(channel string, gossiper mesh.Gossiper) (mesh.Gossip, error)
}

func NewMemDBStore(router Router) (*memDBStore, error) {
	db, err := memdb.NewMemDB(&memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"messages": &memdb.TableSchema{
				Name: "messages",
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
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	s := &memDBStore{
		db:         db,
		topicIndex: TenantTopicIndexer(),
	}
	state, err := state.NewStore("mqtt-topics", s, router)
	if err != nil {
		return nil, err
	}
	s.state = state
	return s, nil
}

type topicIndexer struct {
	root *Node
}

func TenantTopicIndexer() *topicIndexer {
	return &topicIndexer{
		root: NewNode("_root", "_all"),
	}
}

func (t *topicIndexer) Lookup(tenant string, pattern []byte) (RetainedMessageList, error) {
	var vals RetainedMessageList
	topic := NewTopic(pattern)
	t.root.Apply(tenant, topic, func(node *Node) bool {
		if node.Message != nil {
			vals.RetainedMessages = append(vals.RetainedMessages, node.Message)
		}
		return false
	})
	return vals, nil
}

func (s *topicIndexer) Index(message *RetainedMessage) error {
	topic := NewTopic(message.GetTopic())
	node := s.root.Upsert(message.GetTenant(), topic)
	node.Message = message
	return nil
}

func (m *memDBStore) ByID(id string) (*RetainedMessage, error) {
	var res *RetainedMessage
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.first(tx, "id", id)
		return
	})
}
func (m *memDBStore) All() (RetainedMessageList, error) {
	var res RetainedMessageList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "id")
		return
	})
}
func (m *memDBStore) ByTenant(tenant string) (RetainedMessageList, error) {
	var res RetainedMessageList
	return res, m.read(func(tx *memdb.Txn) (err error) {
		res, err = m.all(tx, "tenant", tenant)
		return
	})
}
func (m *memDBStore) ByTopicPattern(tenant string, pattern []byte) (RetainedMessageList, error) {
	return m.topicIndex.Lookup(tenant, pattern)
}
func (m *memDBStore) Create(message *RetainedMessage) error {
	if message.ID == "" {
		id, err := MakeTopicID(message.Tenant, message.Topic)
		if err != nil {
			return err
		}
		message.ID = id
	}
	message.LastAdded = now()
	err := m.topicIndex.Index(message)
	if err != nil {
		return err
	}
	return m.state.Upsert(message)
}
func (m *memDBStore) insert(messages []*RetainedMessage) error {
	return m.write(func(tx *memdb.Txn) error {
		for _, message := range messages {
			if message.IsAdded() {
				err := m.topicIndex.Index(message)
				if err != nil {
					return err
				}
			}
			if message.IsRemoved() {
				message.Payload = nil
				err := m.topicIndex.Index(message)
				if err != nil {
					return err
				}
			}
			err := tx.Insert("messages", message)
			if err != nil {
				return err
			}
		}
		tx.Commit()
		return nil
	})
}