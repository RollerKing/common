package election

import (
	"context"
	"errors"
	"github.com/qjpcpu/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type Role int

const (
	// Leader leader role
	Leader Role = 1
	// Candidate candidate role
	Candidate = 2
)

// HA ha handler
type HA struct {
	endpoints []string
	keyPrefix string
	ttl       int
	last      Role
	roleC     chan Role
}

// New new ha
func New(endpoints []string, key string) *HA {
	return &HA{
		endpoints: endpoints,
		keyPrefix: key,
		last:      Candidate,
		ttl:       30,
		roleC:     make(chan Role, 1),
	}
}

// TTL set ttl
func (h *HA) TTL(ttl int) *HA {
	if ttl > 0 {
		h.ttl = ttl
	}
	return h
}

// IsLeader is leader
func (h *HA) IsLeader() bool {
	return h.last == Leader
}

// GetRole current role
func (h *HA) GetRole() Role {
	return h.last
}

// RoleC get roll channel
func (h *HA) RoleC() <-chan Role {
	return h.roleC
}

// Start start ha
func (h *HA) Start() error {
	if len(h.endpoints) == 0 {
		return errors.New("no endpoints")
	}
	if h.keyPrefix == "" {
		return errors.New("bad key")
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: h.endpoints})
	if err != nil {
		log.Error(err)
		return err
	}
	defer cli.Close()
	for {
		h.startSession(cli)
	}
}

func (h *HA) notifyState(state Role) {
	if h.last != state {
		h.roleC <- state
		h.last = state
		log.Debug("switch to ", state.String())
	}
}

func (h *HA) startSession(cli *clientv3.Client) {
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(h.ttl))
	if err != nil {
		log.Error("create session fail:", err)
		return
	}
	defer session.Close()
	val := "ha"
	elec := concurrency.NewElection(session, h.keyPrefix)
	for {
		if err := elec.Campaign(context.Background(), val); err != nil {
			log.Error("campaign fail:", err)
			h.notifyState(Candidate)
			continue
		}
		h.notifyState(Leader)
		<-session.Done()
		break
	}
	h.notifyState(Candidate)
}

func (r Role) String() string {
	switch r {
	case Leader:
		return "Leader"
	case Candidate:
		return "Candidate"
	default:
		return "Unknown"
	}
}
