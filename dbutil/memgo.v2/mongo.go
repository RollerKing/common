package memgo

import (
	"fmt"
	"github.com/globalsign/mgo"
	"sync"
)

var (
	defaultSession *mgo.Session
	sessionLock    = new(sync.RWMutex)
	sessions       = make(map[string]*mgo.Session)
)

// The seed servers must be provided in the following format:
//
//     [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
//
// For example, it may be as simple as:
//
//     localhost
//
// Or more involved like:
//
//     mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
//
// InitMongo init mongo session with optional alias
func InitMongo(url string, alias ...string) error {
	var name string
	if len(alias) > 0 {
		name = alias[0]
	}
	if _, ok := sessions[name]; ok {
		return fmt.Errorf("duplicate session:[%s]", name)
	}
	s, err := mgo.Dial(url)
	if err != nil {
		return err
	}
	sessionLock.Lock()
	defer sessionLock.Unlock()
	if len(sessions) == 0 {
		defaultSession = s
	}
	sessions[name] = s
	return nil
}

// GetSession get mongo session
func GetSession() *mgo.Session {
	if defaultSession != nil {
		return defaultSession.Copy()
	}
	return nil
}

// GetSessionBy get session by alias
func GetSessionBy(alias string) *mgo.Session {
	sessionLock.RLock()
	defer sessionLock.RUnlock()
	if s, ok := sessions[alias]; ok && s != nil {
		return s.Copy()
	}
	return nil
}
