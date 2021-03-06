package sessions

//go:generate protoc -I${GOPATH}/src -I${GOPATH}/src/github.com/vx-labs/mqtt-broker/sessions/ --go_out=plugins=grpc:. types.proto

type sessionFilter func(Session) bool
type SessionSet []Session

func (set SessionSet) Filter(filters ...sessionFilter) SessionSet {
	copy := make(SessionSet, 0)
	for _, session := range set {
		accepted := true
		for _, f := range filters {
			if !f(session) {
				accepted = false
				break
			}
		}
		if accepted {
			copy = append(copy, session)
		}
	}
	return copy
}
func (set SessionSet) Apply(f func(s Session)) {
	for _, session := range set {
		f(session)
	}
}
func (set SessionSet) ApplyIdx(f func(idx int, s Session)) {
	for idx, session := range set {
		f(idx, session)
	}
}

func (set SessionSet) ApplyE(f func(s Session) error) error {
	for _, session := range set {
		if err := f(session); err != nil {
			return err
		}
	}
	return nil
}
