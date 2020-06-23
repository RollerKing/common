package debug

import (
	"encoding/json"
	"os/user"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"github.com/xujiajun/nutsdb"
	"github.com/xujiajun/nutsdb/ds/zset"
)

const (
	defaultDir = ".disk-cache"
)

type DiskCache struct {
	db         *nutsdb.DB
	bucketSize map[string]int
}

func NewHomeDiskCache() (*DiskCache, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	return New(filepath.Join(usr.HomeDir, defaultDir))
}

func New(dbdir string) (*DiskCache, error) {
	opt := nutsdb.DefaultOptions
	opt.Dir = dbdir
	if db, err := nutsdb.Open(opt); err != nil {
		return nil, err
	} else {
		return &DiskCache{db: db, bucketSize: make(map[string]int)}, nil
	}
}

func (dc *DiskCache) Close() {
	dc.db.Close()
}

func (dc *DiskCache) SetBucketSize(bucket string, size int) *DiskCache {
	dc.bucketSize[bucket] = size
	return dc
}

func (dc *DiskCache) AddItem(bucket string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	now := time.Now().UnixNano()
	return dc.db.Update(
		func(tx *nutsdb.Tx) error {
			if maxCount, ok := dc.bucketSize[bucket]; ok && maxCount > 0 && dc.bucketExist(bucket) {
				if size, _ := tx.ZCard(bucket); size >= maxCount {
					for i := 0; i < size-maxCount+1; i++ {
						tx.ZPopMin(bucket)
					}
				}
			}
			key := data
			if keyer, ok := v.(Keyer); ok {
				key = []byte(keyer.GetKey())
			}
			return tx.ZAdd(bucket, key, float64(now), data)
		})
}

func (dc *DiskCache) bucketExist(bucket string) bool {
	if dc.db.SortedSetIdx == nil {
		return false
	} else if _, ok := dc.db.SortedSetIdx[bucket]; !ok {
		return false
	}
	return true
}

func (dc *DiskCache) ListItem(bucket string, retSlicePtr interface{}) error {
	if !dc.bucketExist(bucket) {
		return nil
	}
	var list []*zset.SortedSetNode
	if err := dc.db.View(
		func(tx *nutsdb.Tx) error {
			if nodes, err := tx.ZMembers(bucket); err != nil {
				return err
			} else {
				for _, node := range nodes {
					list = append(list, node)
				}
			}
			return nil
		}); err != nil {
		return err
	}
	sort.SliceStable(list, func(i, j int) bool {
		return int64(list[i].Score()) > int64(list[j].Score())
	})
	tp := reflect.TypeOf(retSlicePtr)
	v := reflect.ValueOf(retSlicePtr)
	vElem := v.Elem()
	elemType := tp.Elem().Elem()
	var elemIsPtr bool
	if elemIsPtr = elemType.Kind() == reflect.Ptr; elemIsPtr {
		elemType = elemType.Elem()
	}
	for _, node := range list {
		nv := reflect.New(elemType)
		if err := json.Unmarshal(node.Value, nv.Interface()); err != nil {
			continue
		}
		if !elemIsPtr {
			nv = nv.Elem()
		}
		vElem = reflect.Append(vElem, nv)
	}
	v.Elem().Set(vElem)
	return nil
}

type Keyer interface {
	GetKey() string
}
