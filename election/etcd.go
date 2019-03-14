package election

import (
	"context"
	"errors"
	"github.com/qjpcpu/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.etcd.io/etcd/mvcc/mvccpb"
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
		log.Errorf("[election]%v", err)
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
		log.Debugf("[election]switch to %s", state.String())
	}
}

func (h *HA) startSession(cli *clientv3.Client) {
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(h.ttl))
	if err != nil {
		log.Errorf("[election]create session fail:%v", err)
		return
	}
	defer session.Close()
	defer h.notifyState(Candidate)
	val := "election"
	elec := concurrency.NewElection(session, h.keyPrefix)
	for {
		if err := elec.Campaign(context.Background(), val); err != nil {
			log.Errorf("[election]campaign fail:%v", err)
			h.notifyState(Candidate)
			continue
		}
		lderVal, err := elec.Leader(context.Background())
		if err != nil {
			log.Errorf("[election]get leader key fail:%v", err)
			h.notifyState(Candidate)
			continue
		}
		if len(lderVal.Kvs) == 0 {
			log.Error("[election]get empty leader key")
			h.notifyState(Candidate)
			continue
		}
		lderKey := string(lderVal.Kvs[0].Key)
		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		log.Debugf("[election]start watch key %s", lderKey)
		wch := cli.Watch(cctx, lderKey, clientv3.WithRev(lderVal.Header.GetRevision()))
		h.notifyState(Leader)
		for {
			select {
			case <-session.Done():
				return
			case wr := <-wch:
				for _, ev := range wr.Events {
					if ev.Type == mvccpb.DELETE {
						log.Debugf("[election] %s is lost unexpected, so resign myself", lderKey)
						elec.Resign(context.Background())
						return
					}
				}
			}
		}
	}
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
