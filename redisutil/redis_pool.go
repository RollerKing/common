package redisutil

import (
	"github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
	"time"
)

var (
	gRedisPool    *redis.Pool
	gRedisCluster *redisc.Cluster
)

// Options redis options
type Options struct {
	ConnectTimeout int  `json:"connect_timeout" toml:"connect_timeout"`
	ReadTimeout    int  `json:"read_timeout" toml:"read_timeout"`
	WriteTimeout   int  `json:"write_timeout" toml:"write_timeout"`
	MaxActive      int  `json:"max_active" toml:"max_active"`
	MaxIdle        int  `json:"max_idle" toml:"max_idle"`
	IdleTimeout    int  `json:"idle_timeout" toml:"idle_timeout"`
	Wait           bool `json:"wait" toml:"wait"`
}

// OptFunc redis option function
type OptFunc func(*Options)

// RedisInfo redis config
type RedisInfo struct {
	Address  string `json:"address" toml:"address"`
	DB       string `json:"db,omitempty" toml:"db,omitempty"`
	Password string `json:"password,omitempty" toml:"password,omitempty"`
	Options
}

// RedisClusterInfo redis cluster config
type RedisClusterInfo struct {
	StartupNodes []string `json:"startup_nodes" toml:"startup_nodes"`
	DB           string   `json:"db" toml:"db"`
	Password     string   `json:"password,omitempty" toml:"password,omitempty"`
	Options
}

// ConvertToOptFunc convert option to func
func ConvertToOptFunc(opt Options) OptFunc {
	return func(option *Options) {
		if opt.ConnectTimeout != 0 {
			option.ConnectTimeout = opt.ConnectTimeout
		}
		if opt.ReadTimeout != 0 {
			option.ReadTimeout = opt.ReadTimeout
		}
		if opt.WriteTimeout != 0 {
			option.WriteTimeout = opt.WriteTimeout
		}
		if opt.MaxActive != 0 {
			option.MaxActive = opt.MaxActive
		}
		if opt.MaxIdle != 0 {
			option.MaxIdle = opt.MaxIdle
		}
		if opt.IdleTimeout != 0 {
			option.IdleTimeout = opt.IdleTimeout
		}
		option.Wait = opt.Wait
	}
}

// GetPool get redis pool
func GetPool() *redis.Pool {
	return gRedisPool
}

// GetCluster get redis cluster
func GetCluster() *redisc.Cluster {
	return gRedisCluster
}

// CreatePool create redis pool
func CreatePool(conn string, redisDB, passwd string, wrappers ...OptFunc) *redis.Pool {
	var opt = &Options{
		MaxIdle:        200,
		MaxActive:      200,
		IdleTimeout:    2,
		Wait:           false,
		ConnectTimeout: 2,
		ReadTimeout:    2,
		WriteTimeout:   2,
	}
	for _, fn := range wrappers {
		fn(opt)
	}
	return &redis.Pool{
		MaxIdle:     opt.MaxIdle,
		MaxActive:   opt.MaxActive,
		IdleTimeout: time.Duration(opt.IdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", conn, time.Duration(opt.ConnectTimeout)*time.Second, time.Duration(opt.ReadTimeout)*time.Second, time.Duration(opt.WriteTimeout)*time.Second)
			if err != nil {
				return nil, err
			}

			if passwd != "" {
				if _, err := c.Do("AUTH", passwd); err != nil {
					c.Close()
					return nil, err
				}
			}

			if redisDB != "" {
				if _, err = c.Do("SELECT", redisDB); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// InitRedis init default redis pool
func InitRedis(conn string, redisDB, passwd string, optfunc ...OptFunc) {
	gRedisPool = CreatePool(conn, redisDB, passwd, optfunc...)
}

// InitCluster init default redis cluster
func InitCluster(startupNodes []string, db, pwd string, wrappers ...OptFunc) {
	gRedisCluster = CreateCluster(startupNodes, db, pwd, wrappers...)
}

// CreateCluster create redis cluster
func CreateCluster(startupNodes []string, db, pwd string, wrappers ...OptFunc) *redisc.Cluster {
	clusterConfig := &RedisClusterInfo{
		StartupNodes: startupNodes,
		DB:           db,
		Password:     pwd,
	}
	var opt = &Options{
		MaxIdle:        200,
		MaxActive:      200,
		IdleTimeout:    2,
		Wait:           false,
		ConnectTimeout: 2,
		ReadTimeout:    2,
		WriteTimeout:   2,
	}
	for _, fn := range wrappers {
		fn(opt)
	}
	clusterConfig.Options = *opt
	return &redisc.Cluster{
		StartupNodes: clusterConfig.StartupNodes,
		DialOptions:  clusterDialOptions(clusterConfig),
		CreatePool:   clusterCreatePool(clusterConfig),
	}
}

func clusterDialOptions(ci *RedisClusterInfo) []redis.DialOption {
	return []redis.DialOption{
		redis.DialConnectTimeout(time.Duration(ci.ConnectTimeout) * time.Second),
		redis.DialConnectTimeout(time.Duration(ci.ReadTimeout) * time.Second),
		redis.DialWriteTimeout(time.Duration(ci.WriteTimeout) * time.Second),
	}
}

func clusterCreatePool(ci *RedisClusterInfo) func(addr string, opts ...redis.DialOption) (*redis.Pool, error) {
	// set defaults
	if ci.MaxIdle == 0 {
		ci.MaxIdle = 200
	}
	if ci.MaxActive == 0 {
		ci.MaxActive = 200
	}
	if ci.IdleTimeout == 0 {
		ci.IdleTimeout = 2
	}
	if ci.ConnectTimeout == 0 {
		ci.ConnectTimeout = 2
	}
	if ci.ReadTimeout == 0 {
		ci.ReadTimeout = 2
	}
	if ci.WriteTimeout == 0 {
		ci.WriteTimeout = 2
	}
	return func(addr string, opts ...redis.DialOption) (*redis.Pool, error) {
		return &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", addr, opts...)
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			MaxIdle:     ci.MaxIdle,
			MaxActive:   ci.MaxActive,
			IdleTimeout: time.Duration(ci.IdleTimeout) * time.Second,
			Wait:        ci.Wait,
		}, nil
	}
}
