package memgo

import (
	"github.com/globalsign/mgo/bson"
	"github.com/qjpcpu/common/json"
)

// RawJSON raw json and store as bson
type RawJSON []byte

func (r *RawJSON) SetBSON(b bson.Raw) error {
	var i interface{}
	if err := b.Unmarshal(&i); err != nil {
		return err
	}
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	*r = RawJSON(data)
	return nil
}

func (r *RawJSON) GetBSON() (interface{}, error) {
	if r == nil || len(*r) == 0 {
		return nil, nil
	}
	var i interface{}
	if err := json.Unmarshal([]byte(*r), &i); err != nil {
		return nil, err
	}
	return i, nil
}

func (r *RawJSON) UnmarshalJSON(b []byte) error {
	*r = RawJSON(b)
	return nil
}

func (r RawJSON) MarshalJSON() ([]byte, error) {
	rj := json.RawMessage(r)
	return rj.MarshalJSON()
}

func (r RawJSON) Bytes() []byte {
	return []byte(r)
}

func (r RawJSON) String() string {
	return string(r)
}
