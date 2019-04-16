package subscriptions

import (
	"errors"

	memdb "github.com/hashicorp/go-memdb"
)

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
	data, err := tx.First(table, index, value...)
	if err != nil {
		return res, err
	}
	wrapper, ok := data.(subscriptionWrapper)
	if !ok {
		return res, errors.New("invalid type fetched")
	}
	res = wrapper.pb
	return res, nil
}
func (m *memDBStore) all(tx *memdb.Txn, index string, value ...interface{}) (SubscriptionList, error) {
	var set SubscriptionList
	iterator, err := tx.Get(table, index, value...)
	if err != nil {
		return set, err
	}
	for {
		data := iterator.Next()
		if data == nil {
			return set, nil
		}
		res, ok := data.(subscriptionWrapper)
		if !ok {
			return set, errors.New("invalid type fetched")
		}
		if res.pb.IsAdded() {
			set.Subscriptions = append(set.Subscriptions, res.pb)
		}
	}
}
