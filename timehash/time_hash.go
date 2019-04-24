package timehash

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"time"
)

/*
 数据存储的样例: key: {"data":YOUR_DATA,"comm":{"exp":1487311773,"ver": 2}}
 data存放用户数据,comm存放内置数据
 comm.ver: 每SET一次数据,ver+1
 comm.exp: 过期unix时间戳
*/
var ErrNoCache = errors.New("No such data")

const reserveField = "__expire_at_max__"

// redis-cli --eval h.lua h , c current_timestamp  isCut
var getScript = redis.NewScript(1, `
if redis.call("HEXISTS", KEYS[1],ARGV[1]) == 0 then
  return nil
end
local payload = redis.call("HGET",KEYS[1],ARGV[1])
local data = cjson.decode(payload)
if tonumber(data["comm"]["exp"]) < tonumber(ARGV[2]) then
  redis.call("HDEL",KEYS[1],ARGV[1])
  return nil
else
  if tonumber(ARGV[3]) == 1 then
    redis.call("HDEL",KEYS[1],ARGV[1])
  end
  return payload
end
`)

// redis-cli --eval h.lua h , c  'data' current_timestamp
var setScript = redis.NewScript(1, `
if redis.call("HEXISTS", KEYS[1],ARGV[1]) == 1 then
  local data = cjson.decode(ARGV[2])
  local old_data = cjson.decode(redis.call("HGET",KEYS[1],ARGV[1]))
  if old_data["comm"]["ver"] == nil then old_data["comm"]["ver"] = 0 end
  if old_data["comm"]["exp"] == nil or tonumber(old_data["comm"]["exp"]) < tonumber(ARGV[3]) then old_data["comm"]["ver"] = -1 end
  data["comm"]["ver"] = old_data["comm"]["ver"] + 1
  ARGV[2] = cjson.encode(data)
end
redis.call("HSET",KEYS[1],ARGV[1],ARGV[2])
local decoded = cjson.decode(ARGV[2])
local exp = tonumber(decoded["comm"]["exp"])
local exp_key = "__expire_at_max__"
if redis.call("HEXISTS", KEYS[1],exp_key) == 1 then
  local oexp = tonumber(redis.call("HGET",KEYS[1],exp_key))
  if oexp < exp then
    redis.call("HSET",KEYS[1],exp_key,exp)
    redis.call("EXPIRE",KEYS[1],exp - ARGV[3])
  end
else
  redis.call("HSET",KEYS[1],exp_key,exp)
  redis.call("EXPIRE",KEYS[1],exp - ARGV[3])
end
return tonumber(decoded["comm"]["ver"])
`)

type CommonPayload struct {
	ExpireAt int64 `json:"exp"`
	Version  int64 `json:"ver"`
}

type cachePayloadIn struct {
	CommonPayload `json:"comm"`
	Data          interface{} `json:"data"`
}

type cachePayloadOut struct {
	CommonPayload `json:"comm"`
	Data          *json.RawMessage `json:"data"`
}

func Del(conn redis.Conn, key string, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}
	params := []interface{}{key}
	for _, key := range fields {
		params = append(params, key)
	}
	_, err := conn.Do("HDEL", params...)
	return err
}

func Set(conn redis.Conn, key, filed string, data interface{}, ttl int) (int64, error) {
	var expirtAt time.Time
	if ttl > 0 {
		expirtAt = time.Now().Add(time.Duration(ttl) * time.Second)
	} else {
		expirtAt = time.Now().AddDate(10, 0, 0)
	}
	payload := cachePayloadIn{
		Data: data,
		CommonPayload: CommonPayload{
			ExpireAt: expirtAt.Unix(),
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	ver, err := redis.Int64(setScript.Do(conn, key, filed, b, time.Now().Unix()))
	return ver, err
}

// data 必须为指针
func fetch(conn redis.Conn, key, field string, data interface{}, isCut bool) (int64, error) {
	cut := 0
	if isCut {
		cut = 1
	}
	res, err := redis.Bytes(getScript.Do(conn, key, field, time.Now().Unix(), cut))
	if err != nil {
		if err == redis.ErrNil {
			return 0, ErrNoCache
		}
		return 0, err
	}
	p := cachePayloadOut{}
	if err = json.Unmarshal(res, &p); err != nil {
		return 0, err
	}
	return p.CommonPayload.Version, json.Unmarshal(*p.Data, data)
}

// data 必须为指针
func Get(conn redis.Conn, key, field string, data interface{}) (int64, error) {
	return fetch(conn, key, field, data, false)
}

// data 必须为指针
func Cut(conn redis.Conn, key, field string, data interface{}) (int64, error) {
	return fetch(conn, key, field, data, true)
}

func GetAll(conn redis.Conn, key string) (map[string][]byte, error) {
	res := make(map[string][]byte)
	m, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return res, err
	}
	now := time.Now().Unix()
	for k, v := range m {
		if k == reserveField || len(v) == 0 {
			continue
		}
		p := cachePayloadOut{}
		if err = json.Unmarshal([]byte(v), &p); err != nil {
			return res, err
		}
		if p.Data != nil && p.ExpireAt > now {
			res[k] = []byte(*p.Data)
		}
	}
	return res, nil
}
