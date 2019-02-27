package memgo

import (
	"errors"
	"gopkg.in/mgo.v2"
	"sync"
	"time"
)

var (
	defaultSession *mgo.Session
	sessionLock    = new(sync.RWMutex)
	sessions       = make(map[string]*mgo.Session)
)

// MongoConfig config
type MongoConfig struct {
	Addrs          []string `json:"addrs" toml:"addrs" example:"127.0.0.1:27017"`
	Source         string   `json:"source" toml:"source" example:""`
	ReplicaSetName string   `json:"replica_set_name" toml:"replica_set_name" example:""`
	Timeout        int      `json:"timeout" toml:"timeout" example:"2"`
	Username       string   `json:"username" toml:"username" example:""`
	Password       string   `json:"password" toml:"password" example:""`
	Mode           *int     `json:"mode,omitempty" toml:"mode,omitempty" example:"3"`
	Alias          string   `json:"alias" toml:"alias" example:"default"`
}

func (mc *MongoConfig) getMongoDialInfo() *mgo.DialInfo {
	if mc.Mode == nil {
		m := int(mgo.PrimaryPreferred)
		mc.Mode = &m
	}
	if *mc.Mode > int(mgo.Nearest) {
		m := int(mgo.PrimaryPreferred)
		mc.Mode = &m
	}
	if mc.Timeout == 0 {
		mc.Timeout = 2
	}
	return &mgo.DialInfo{
		Addrs:          mc.Addrs,
		Source:         mc.Source,
		ReplicaSetName: mc.ReplicaSetName,
		Timeout:        time.Duration(mc.Timeout) * time.Second,
		Username:       mc.Username,
		Password:       mc.Password,
	}
}

// InitMongo init mongo session
func InitMongo(mc *MongoConfig) error {
	dialInfo := mc.getMongoDialInfo()
	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return err
	}
	s.SetMode(mgo.Mode(*mc.Mode), true)
	sessionLock.Lock()
	defer sessionLock.Unlock()
	if _, ok := sessions[mc.Alias]; ok {
		return errors.New("duplicate session:" + mc.Alias)
	}
	if len(sessions) == 0 {
		defaultSession = s
	}
	sessions[mc.Alias] = s
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
