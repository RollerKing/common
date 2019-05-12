package election

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync/atomic"

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
	stopC     chan struct{}
	isStopped int32
	logger    Logger
}

// New new ha
func New(endpoints []string, key string) *HA {
	return &HA{
		endpoints: endpoints,
		keyPrefix: key,
		last:      Candidate,
		ttl:       30,
		roleC:     make(chan Role, 1),
		stopC:     make(chan struct{}, 1),
		logger:    NullLogger{},
	}
}

// SetLogger logger
func (h *HA) SetLogger(l Logger) *HA {
	if l != nil {
		h.logger = l
	}
	return h
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
		h.logger.Errorf("[election]%v", err)
		return err
	}
	defer cli.Close()
	for h.isStopped == 0 {
		h.startSession(cli)
	}
	return nil
}

// Stop ha
func (h *HA) Stop() {
	if atomic.CompareAndSwapInt32(&h.isStopped, 0, 1) {
		close(h.stopC)
	}
}

func (h *HA) notifyState(state Role) {
	if h.last != state {
		h.roleC <- state
		h.last = state
		h.logger.Debugf("[election]switch to %s", state.String())
	}
}

func (h *HA) startSession(cli *clientv3.Client) {
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(h.ttl))
	if err != nil {
		h.logger.Errorf("[election]create session fail:%v", err)
		return
	}
	defer session.Close()
	defer h.notifyState(Candidate)
	val := "election"
	if h, err := os.Hostname(); err == nil {
		val += fmt.Sprintf(":%s:%d", h, os.Getpid())
	}
	elec := concurrency.NewElection(session, h.keyPrefix)
	for {
		if err := elec.Campaign(context.Background(), val); err != nil {
			h.logger.Errorf("[election]campaign fail:%v", err)
			h.notifyState(Candidate)
			continue
		}
		lderVal, err := elec.Leader(context.Background())
		if err != nil {
			h.logger.Errorf("[election]get leader key fail:%v", err)
			h.notifyState(Candidate)
			continue
		}
		if len(lderVal.Kvs) == 0 {
			h.logger.Error("[election]get empty leader key")
			h.notifyState(Candidate)
			continue
		}
		lderKey := string(lderVal.Kvs[0].Key)
		cctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		h.logger.Debugf("[election]start watch key %s", lderKey)
		wch := cli.Watch(cctx, lderKey, clientv3.WithRev(lderVal.Header.GetRevision()))
		h.notifyState(Leader)
		for {
			select {
			case <-h.stopC:
				elec.Resign(context.Background())
				return
			case <-session.Done():
				return
			case wr := <-wch:
				for _, ev := range wr.Events {
					if ev.Type == mvccpb.DELETE {
						h.logger.Debugf("[election] %s is lost unexpected, so resign myself", lderKey)
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

type Logger interface {
	Debugf(string, ...interface{})
	Debug(...interface{})
	Infof(string, ...interface{})
	Info(...interface{})
	Errorf(string, ...interface{})
	Error(...interface{})
}

// NullLogger drop all log
type NullLogger struct{}

func (null NullLogger) Debugf(f string, v ...interface{}) {}
func (null NullLogger) Debug(v ...interface{})            {}
func (null NullLogger) Infof(f string, v ...interface{})  {}
func (null NullLogger) Info(v ...interface{})             {}
func (null NullLogger) Errorf(f string, v ...interface{}) {}
func (null NullLogger) Error(v ...interface{})            {}

// StdLogger stderr log
type StdLogger struct{}

func (sl StdLogger) Debugf(f string, v ...interface{}) {
	log.Printf(f+"\n", v...)
}
func (sl StdLogger) Debug(v ...interface{}) {
	log.Println(v...)
}
func (sl StdLogger) Infof(f string, v ...interface{}) {
	log.Printf(f+"\n", v...)
}
func (sl StdLogger) Info(v ...interface{}) {
	log.Println(v...)
}
func (sl StdLogger) Errorf(f string, v ...interface{}) {
	log.Printf(f+"\n", v...)
}
func (sl StdLogger) Error(v ...interface{}) {
	log.Println(v...)
}
