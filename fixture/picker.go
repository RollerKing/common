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

// Visitor func
type Visitor func(path string, tp reflect.Type, v interface{}) (isContinue bool)

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
	walkVal([]string{}, v.Type(), v, visitOnce(pathHitterToVisitor(pathFn, vals)))
	return
}

// Walk object
func Walk(obj interface{}, visitFn Visitor) {
	if obj == nil {
		return
	}
	v := reflect.ValueOf(obj)
	walkVal([]string{}, v.Type(), v, visitOnce(visitFn))
}

// WalkLeaf call visitFn only when primitive tyeps
func WalkLeaf(obj interface{}, visitFn Visitor) {
	fn := func(path string, tp reflect.Type, v interface{}) bool {
		if IsPrimitiveType(tp) || IsPrimitivePtrType(tp) || IsTimePtrType(tp) || IsTimeType(tp) {
			return visitFn(path, tp, v)
		}
		return true
	}
	Walk(obj, fn)
}

func visitOnce(visit Visitor) Visitor {
	onceMap := make(map[string]bool)
	return func(path string, tp reflect.Type, v interface{}) bool {
		if _, ok := onceMap[path]; ok {
			return true
		}
		onceMap[path] = true
		return visit(path, tp, v)
	}
}

func pathHitterToVisitor(pathFn PathHitter, vals Values) Visitor {
	return func(path string, tp reflect.Type, v interface{}) bool {
		if pathFn(path) {
			vals.setVal(path, v)
		}
		return true
	}
}

func walkVal(steps []string, t reflect.Type, v reflect.Value, visit Visitor) bool {
	path := buildPath(steps)
	switch t.Kind() {
	case reflect.String:
		if !visit(path, t, v.String()) {
			return false
		}
	case reflect.Bool:
		if !visit(path, t, v.Bool()) {
			return false
		}
	case reflect.Int64:
		if !visit(path, t, int64(v.Int())) {
			return false
		}
	case reflect.Int:
		if !visit(path, t, int(v.Int())) {
			return false
		}
	case reflect.Int8:
		if !visit(path, t, int8(v.Int())) {
			return false
		}
	case reflect.Int16:
		if !visit(path, t, int16(v.Int())) {
			return false
		}
	case reflect.Int32:
		if !visit(path, t, int32(v.Int())) {
			return false
		}
	case reflect.Uint:
		if !visit(path, t, v.Uint()) {
			return false
		}
	case reflect.Uint8:
		if !visit(path, t, uint8(v.Uint())) {
			return false
		}
	case reflect.Uint16:
		if !visit(path, t, uint16(v.Uint())) {
			return false
		}
	case reflect.Uint32:
		if !visit(path, t, uint32(v.Uint())) {
			return false
		}
	case reflect.Uint64:
		if !visit(path, t, uint64(v.Uint())) {
			return false
		}
	case reflect.Uintptr:
		if !visit(path, t, uintptr(v.Uint())) {
			return false
		}
	case reflect.Float32:
		if !visit(path, t, float32(v.Float())) {
			return false
		}
	case reflect.Float64:
		if !visit(path, t, v.Float()) {
			return false
		}
	case reflect.Struct:
		if !visit(path, t, v.Interface()) {
			return false
		}
		walkStruct(steps, t, v, visit)
	case reflect.Ptr:
		if !visit(path, t, v.Interface()) {
			return false
		}
		if !v.IsNil() {
			if !walkVal(steps, t.Elem(), v.Elem(), visit) {
				return false
			}
		}
	case reflect.Map:
		if !visit(path, t, v.Interface()) {
			return false
		}
		if !v.IsNil() {
			if !walkMap(steps, t.Key(), t.Elem(), v, visit) {
				return false
			}
		}
	case reflect.Slice, reflect.Array:
		if !visit(path, t, v.Interface()) {
			return false
		}
		if !v.IsNil() {
			if !walkSlice(steps, t.Elem(), v, visit) {
				return false
			}
		}
	case reflect.Chan:
		return false
	case reflect.Interface:
		if !visit(path, t, v.Interface()) {
			return false
		}
		if !v.IsNil() {
			var isContinue bool
			if v.Elem().Kind() == reflect.Ptr {
				isContinue = walkVal(steps, v.Elem().Elem().Type(), v.Elem().Elem(), visit)
			} else {
				isContinue = walkVal(steps, v.Elem().Type(), v.Elem(), visit)
			}
			return isContinue
		}
	}
	return true
}

func walkMap(steps []string, kt, vt reflect.Type, v reflect.Value, fn Visitor) bool {
	keys := v.MapKeys()
	for _, key := range keys {
		vv := v.MapIndex(key)
		if !walkVal(append(steps, key.String()), vv.Type(), vv, fn) {
			return false
		}
	}
	return true
}

func walkStruct(steps []string, t reflect.Type, v reflect.Value, fn Visitor) bool {
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)
		if !isExported(ft.Name) {
			continue
		}
		if !walkVal(append(steps, ft.Name), ft.Type, fv, fn) {
			return false
		}
	}
	return true
}

func walkSlice(steps []string, et reflect.Type, v reflect.Value, fn Visitor) bool {
	for i := 0; i < v.Len(); i++ {
		if !walkVal(appendStep(steps, "[", strconv.FormatInt(int64(i), 10), "]"), et, v.Index(i), fn) {
			return false
		}
	}
	return true
}
