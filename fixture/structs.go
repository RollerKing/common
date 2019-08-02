package fixture

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// option for fill
type option struct {
	MaxLevel        int
	MaxSliceLen     int
	MaxMapLen       int
	PathToValueFunc PathToValueFunc
}

// OptionFunc option
type OptionFunc func(*option)

// PathToValueFunc path to value
type PathToValueFunc func(string, reflect.Type) (interface{}, bool)

// MapKey of map mark
const MapKey = "$$KEY"

// FillStruct fill struct, obj must be pointer
func FillStruct(obj interface{}, optF ...OptionFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fill struct: %v", r)
		}
	}()
	stype := reflect.TypeOf(obj)
	if stype.Kind() != reflect.Ptr {
		return errors.New("should be pointer")
	}
	opt := defaultOpt()
	for _, fn := range optF {
		fn(&opt)
	}
	f := &filler{option: &opt}
	f.initializeVal([]string{}, stype.Elem(), reflect.ValueOf(obj).Elem(), opt.MaxLevel)
	return
}

// SetMaxLevel dfs depth
func SetMaxLevel(lvl int) OptionFunc {
	return func(opt *option) {
		opt.MaxLevel = lvl
	}
}

// SetMaxSliceLen slice size
func SetMaxSliceLen(size int) OptionFunc {
	return func(opt *option) {
		opt.MaxSliceLen = size
	}
}

// SetMaxMapLen map size
func SetMaxMapLen(size int) OptionFunc {
	return func(opt *option) {
		opt.MaxMapLen = size
	}
}

// SetPathToValueFunc customize path to value function
func SetPathToValueFunc(fn PathToValueFunc) OptionFunc {
	return func(opt *option) {
		opt.PathToValueFunc = fn
	}
}

// WithSysPVFunc disclose
func WithSysPVFunc(fn PathToValueFunc) OptionFunc {
	return SetPathToValueFunc(func(p string, t reflect.Type) (interface{}, bool) {
		v, ok := fn(p, t)
		if !ok {
			return defaultPathToValueFunc(p, t)
		}
		return v, ok
	})
}

func (fl *filler) initializeStruct(steps []string, t reflect.Type, v reflect.Value, level int) {
	if level < 0 {
		return
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)
		if !isExported(ft.Name) {
			continue
		}
		if level-1 < 0 {
			break
		}
		var offset int
		if ft.Anonymous {
			offset = 1
		}
		if ft.Type.Kind() == reflect.Ptr {
			fv := reflect.New(ft.Type.Elem())
			fl.initializeVal(appendStep(steps, ft.Name), ft.Type.Elem(), fv.Elem(), level-1+offset)
			f.Set(fv)
		} else {
			fl.initializeVal(appendStep(steps, ft.Name), ft.Type, f, level-1+offset)
		}
	}
}
func appendStep(steps []string, stepArgs ...interface{}) []string {
	var step string
	for _, arg := range stepArgs {
		step += fmt.Sprint(arg)
	}
	return append(steps, step)
}

func removeString(list []string, str string) []string {
	offset := 0
	for i, ele := range list {
		if ele == str {
			offset++
		} else if offset > 0 {
			list[i-offset] = ele
		}
	}
	return list[:len(list)-offset]
}

func (fl *filler) initializeSlice(steps []string, t reflect.Type, elemt reflect.Type, level int) reflect.Value {
	size := fl.MaxSliceLen
	slicev := reflect.MakeSlice(t, size, size)
	if level < 0 {
		return slicev
	}
	if elemt.Kind() == reflect.Ptr {
		for i := 0; i < size; i++ {
			ele := reflect.New(elemt.Elem())
			fl.initializeVal(appendStep(steps, "[", i, "]"), ele.Elem().Type(), ele.Elem(), level)
			slicev.Index(i).Set(ele)
		}
	} else {
		for i := 0; i < size; i++ {
			ele := reflect.New(elemt)
			fl.initializeVal(appendStep(steps, "[", i, "]"), elemt, ele.Elem(), level)
			slicev.Index(i).Set(ele.Elem())
		}
	}
	return slicev
}

func (fl *filler) initializeMap(steps []string, tk, tv reflect.Type, mapv reflect.Value, level int) {
	if level <= 0 {
		return
	}
	for i := 0; i < fl.MaxMapLen; i++ {
		//key
		var key, val reflect.Value
		if tk.Kind() == reflect.Ptr {
			kptr := reflect.New(tk.Elem())
			fl.initializeVal(appendStep(steps, MapKey), kptr.Elem().Type(), kptr.Elem(), level-1)
			key = kptr
		} else {
			kptr := reflect.New(tk)
			fl.initializeVal(appendStep(steps, MapKey), tk, kptr.Elem(), level-1)
			key = kptr.Elem()
		}
		// value
		if tv.Kind() == reflect.Ptr {
			vptr := reflect.New(tv.Elem())
			fl.initializeVal(appendStep(steps, key.String()), vptr.Elem().Type(), vptr.Elem(), level-1)
			val = vptr
		} else {
			vptr := reflect.New(tv)
			fl.initializeVal(appendStep(steps, key.String()), tv, vptr.Elem(), level-1)
			val = vptr.Elem()
		}
		mapv.SetMapIndex(key, val)
	}
}

func (fl *filler) getPathValue(steps []string, tp reflect.Type) (reflect.Value, bool) {
	path := strings.Join(steps, ".")
	if fl.PathToValueFunc == nil {
		return reflect.Value{}, false
	}
	obj, ok := fl.PathToValueFunc(path, tp)
	if !ok {
		return reflect.Value{}, false
	}
	if v := reflect.ValueOf(obj); v.Kind() == reflect.Ptr {
		return v.Elem(), true
	} else {
		return v, true
	}
}

func (fl *filler) initializeVal(steps []string, t reflect.Type, v reflect.Value, level int) {
	if level < 0 {
		return
	}
	switch t.Kind() {
	case reflect.String:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetString(vv.String())
		} else {
			v.SetString(randomString())
		}
	case reflect.Bool:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetBool(vv.Bool())
		} else {
			b := rand.Intn(100)%2 == 0
			v.SetBool(b)
		}
	case reflect.Int:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetInt(vv.Int())
		} else {
			v.SetInt(rand.Int63n(10000))
		}
	case reflect.Int16:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetInt(vv.Int())
		} else {
			v.SetInt(rand.Int63n(16))
		}
	case reflect.Int32:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetInt(vv.Int())
		} else {
			v.SetInt(rand.Int63n(32))
		}
	case reflect.Int64:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetInt(vv.Int())
		} else {
			v.SetInt(rand.Int63n(1000))
		}
	case reflect.Int8:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetInt(vv.Int())
		} else {
			v.SetInt(rand.Int63n(8))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetUint(vv.Uint())
		} else {
			v.SetUint(rand.Uint64() % 100)
		}
	case reflect.Float32, reflect.Float64:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.SetFloat(vv.Float())
		} else {
			v.SetFloat(rand.Float64())
		}
	case reflect.Struct:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.Set(vv)
		} else {
			if t.String() == "time.Time" {
				v.Set(reflect.ValueOf(time.Now()))
			} else {
				fl.initializeStruct(steps, t, v, level)
			}
		}
	case reflect.Ptr:
		fv := reflect.New(t)
		fl.initializeVal(steps, t.Elem(), fv.Elem(), level)
		v.Set(fv)
	case reflect.Map:
		if level >= 0 {
			hash := reflect.MakeMap(t)
			if vv, ok := fl.getPathValue(steps, t); ok {
				v.Set(vv)
			} else {
				fl.initializeMap(steps, t.Key(), t.Elem(), hash, level)
				v.Set(hash)
			}
		}
	case reflect.Slice:
		if level >= 0 {
			array := fl.initializeSlice(steps, t, v.Type().Elem(), level)
			v.Set(array)
		}
	case reflect.Chan:
		v.Set(reflect.MakeChan(t, 0))
	case reflect.Interface:
		if vv, ok := fl.getPathValue(steps, t); ok {
			v.Set(vv)
		} else {
			v.Set(reflect.ValueOf("DYNAMIC"))
		}
	}
}

/*
 random helper
*/
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// REmail random email
func REmail() string {
	return randomString() + "@fixture.com"
}

// RMobile random mobile
func RMobile() string {
	data := md5.Sum([]byte(fmt.Sprintf("%v:%v:%v", time.Now(), time.Now().Nanosecond(), r.Float32())))
	m := make([]byte, 0, 11)
	m = append(m, '1')
	if r.Uint32()&1 != 0 {
		m = append(m, '3')
	} else {
		m = append(m, '5')
	}
	for i := 0; i < 9; i++ {
		m = append(m, data[i]%10+'0')
	}
	return string(m)

}

// RLink random link
func RLink() string {
	l := "http://www.fixture.com"
	size := r.Int()%4 + 1
	for i := 0; i < size; i++ {
		l += "/" + randomString()
	}
	return l
}

// RNumber random number
func RNumber(left, right int64) int64 {
	return left + r.Int63n(right-left)
}

// RTimestamp random unix timestamp
func RTimestamp() int64 {
	return time.Now().Unix()
}

// RString random string
func RString() string {
	return randomString()
}

func randomString() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v:%v:%v", time.Now(), time.Now().Nanosecond(), r.Float32()))))[:8]
}

// IsIntegerType is integer
func IsIntegerType(tp reflect.Type) bool {
	switch tp.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return true
	}
	return false
}

// IsUnsignIntegerType is unsign integer
func IsUnsignIntegerType(tp reflect.Type) bool {
	switch tp.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

// IsFloatType is float
func IsFloatType(tp reflect.Type) bool {
	return tp.Kind() == reflect.Float32 || tp.Kind() == reflect.Float64
}

// IsTimeType is time.Time
func IsTimeType(tp reflect.Type) bool {
	return tp == reflect.TypeOf(time.Time{})
}

// IsStringType is string
func IsStringType(tp reflect.Type) bool {
	return tp.Kind() == reflect.String
}

func defaultPathToValueFunc(path string, tp reflect.Type) (interface{}, bool) {
	list := strings.Split(path, ".")
	finalNode := strings.ToLower(list[len(list)-1])
	switch {
	case strings.Contains(finalNode, "email"):
		return REmail(), true
	case strings.Contains(finalNode, "link") || strings.Contains(finalNode, "url"):
		return RLink(), true
	case strings.Contains(finalNode, "id"):
		if IsIntegerType(tp) {
			return RNumber(time.Now().AddDate(0, -1, 0).Unix(), time.Now().Unix()), true
		} else if IsStringType(tp) {
			return RString(), true
		}
	case finalNode == "status":
		if IsIntegerType(tp) {
			return 1, true
		} else if IsUnsignIntegerType(tp) {
			return uint(1), true
		}
	case strings.Contains(finalNode, "num") && IsIntegerType(tp):
		return RNumber(1, 1000), true
	case strings.Contains(finalNode, "phone"):
		return RMobile(), true
	case strings.Contains(finalNode, "mobile"):
		return RMobile(), true
	case strings.Contains(finalNode, "time"):
		if IsTimeType(tp) {
			return time.Now(), true
		} else if IsIntegerType(tp) {
			return RNumber(time.Now().AddDate(0, -1, 0).Unix(), time.Now().Unix()), true
		}
	}
	return "", false
}

func defaultOpt() option {
	return option{
		MaxLevel:        10,
		MaxMapLen:       2,
		MaxSliceLen:     3,
		PathToValueFunc: defaultPathToValueFunc,
	}
}

type filler struct {
	*option
}

func isExported(fieldName string) bool {
	return len(fieldName) > 0 && (fieldName[0] >= 'A' && fieldName[0] <= 'Z')
}
