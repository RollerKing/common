package fixture

import (
	"errors"
	"reflect"
	"strconv"
	"time"
)

var (
	ErrValueNotExist = errors.New("value not exists")
)

// PathHitter return true if you want get the value of certain path
type PathHitter func(string) bool

type value struct {
	v   interface{}
	err error
}

// Values store pick result
type Values map[string]value

func (v Values) setError(path string, err error) {
	v[path] = value{err: err}
}

func (v Values) setVal(path string, val interface{}) {
	if _, ok := v[path]; !ok {
		v[path] = value{v: val}
	}
}

// Paths of results
func (v Values) Paths() []string {
	var list []string
	for k := range v {
		list = append(list, k)
	}
	return list
}

// Get value of the path
func (v Values) Get(path string) (interface{}, error) {
	val, ok := v[path]
	if !ok {
		return nil, ErrValueNotExist
	}
	return val.v, val.err
}

func (v Values) MustGet(path string) interface{} {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv
}

func (v Values) MustGetString(path string) string {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(string)
}

func (v Values) MustGetStringPtr(path string) *string {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(*string)
}

func (v Values) MustGetInt64(path string) int64 {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(int64)
}

func (v Values) MustGetInt64Ptr(path string) *int64 {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(*int64)
}

func (v Values) MustGetInt(path string) int {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(int)
}

func (v Values) MustGetIntPtr(path string) *int {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(*int)
}

func (v Values) MustGetUint64(path string) uint64 {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(uint64)
}

func (v Values) MustGetUint64Ptr(path string) *uint64 {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(*uint64)
}

func (v Values) MustGetTime(path string) time.Time {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(time.Time)
}

func (v Values) MustGetTimePtr(path string) *time.Time {
	vv, err := v.Get(path)
	if err != nil {
		panic(err)
	}
	return vv.(*time.Time)
}

func newValues() Values {
	return make(Values)
}

// PickValuesByLastNode pick by last field name
func PickValuesByLastNode(obj interface{}, fields ...string) Values {
	fieldsMap := make(map[string]bool)
	for _, f := range fields {
		fieldsMap[f] = true
	}
	fn := func(p string) bool {
		return fieldsMap[LastNodeOfPath(p)]
	}
	return PickValues(obj, fn)
}

// PickValuesByPath pick by full path
func PickValuesByPath(obj interface{}, paths ...string) Values {
	fieldsMap := make(map[string]bool)
	for _, f := range paths {
		fieldsMap[f] = true
	}
	fn := func(p string) bool {
		return fieldsMap[p]
	}
	return PickValues(obj, fn)
}

// PickValues pick by path function
func PickValues(obj interface{}, pathFn PathHitter) (vals Values) {
	vals = newValues()
	if obj == nil {
		return
	}
	v := reflect.ValueOf(obj)
	pickVal([]string{}, v.Type(), v, pathFn, vals)
	return
}

func pickVal(steps []string, t reflect.Type, v reflect.Value, pathFn PathHitter, res Values) {
	path := buildPath(steps)
	switch t.Kind() {
	case reflect.String:
		if pathFn(path) {
			res.setVal(path, v.String())
		}
	case reflect.Bool:
		if pathFn(path) {
			res.setVal(path, v.Bool())
		}
	case reflect.Int64:
		if pathFn(path) {
			res.setVal(path, int64(v.Int()))
		}
	case reflect.Int:
		if pathFn(path) {
			res.setVal(path, int(v.Int()))
		}
	case reflect.Int8:
		if pathFn(path) {
			res.setVal(path, int8(v.Int()))
		}
	case reflect.Int16:
		if pathFn(path) {
			res.setVal(path, int16(v.Int()))
		}
	case reflect.Int32:
		if pathFn(path) {
			res.setVal(path, int32(v.Int()))
		}
	case reflect.Uint:
		if pathFn(path) {
			res.setVal(path, v.Uint())
		}
	case reflect.Uint8:
		if pathFn(path) {
			res.setVal(path, uint8(v.Uint()))
		}
	case reflect.Uint16:
		if pathFn(path) {
			res.setVal(path, uint16(v.Uint()))
		}
	case reflect.Uint32:
		if pathFn(path) {
			res.setVal(path, uint32(v.Uint()))
		}
	case reflect.Uint64:
		if pathFn(path) {
			res.setVal(path, uint64(v.Uint()))
		}
	case reflect.Uintptr:
		if pathFn(path) {
			res.setVal(path, uintptr(v.Uint()))
		}
	case reflect.Float32:
		if pathFn(path) {
			res.setVal(path, float32(v.Float()))
		}
	case reflect.Float64:
		if pathFn(path) {
			res.setVal(path, v.Float())
		}
	case reflect.Struct:
		if pathFn(path) {
			res.setVal(path, v.Interface())
		}
		pickStruct(steps, t, v, pathFn, res)
	case reflect.Ptr:
		if pathFn(path) {
			res.setVal(path, v.Interface())
		}
		if !v.IsNil() {
			pickVal(steps, t.Elem(), v.Elem(), pathFn, res)
		}
	case reflect.Map:
		if pathFn(path) {
			res.setVal(path, v.Interface())
		}
		if !v.IsNil() {
			pickMap(steps, t.Key(), t.Elem(), v, pathFn, res)
		}
	case reflect.Slice, reflect.Array:
		if pathFn(path) {
			res.setVal(path, v.Interface())
		}
		if !v.IsNil() {
			pickSlice(steps, t.Elem(), v, pathFn, res)
		}
	case reflect.Chan:

	case reflect.Interface:
		if pathFn(path) {
			res.setVal(path, v.Interface())
		}
		if !v.IsNil() {
			if v.Elem().Kind() == reflect.Ptr {
				pickVal(steps, v.Elem().Elem().Type(), v.Elem().Elem(), pathFn, res)
			} else {
				pickVal(steps, v.Elem().Type(), v.Elem(), pathFn, res)
			}
		}
	}
}

func pickMap(steps []string, kt, vt reflect.Type, v reflect.Value, fn PathHitter, res Values) {
	keys := v.MapKeys()
	for _, key := range keys {
		vv := v.MapIndex(key)
		pickVal(append(steps, key.String()), vv.Type(), vv, fn, res)
	}
}

func pickStruct(steps []string, t reflect.Type, v reflect.Value, fn PathHitter, res Values) {
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)
		if !isExported(ft.Name) {
			continue
		}
		pickVal(append(steps, ft.Name), ft.Type, fv, fn, res)
	}
}

func pickSlice(steps []string, et reflect.Type, v reflect.Value, fn PathHitter, res Values) {
	for i := 0; i < v.Len(); i++ {
		pickVal(appendStep(steps, "[", strconv.FormatInt(int64(i), 10), "]"), et, v.Index(i), fn, res)
	}
}
