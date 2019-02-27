package fixture

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	ExampleTag = "example"
)

// FillStruct 填充结构体,obj必须为指针
func FillStruct(obj interface{}) error {
	stype := reflect.TypeOf(obj)
	if stype.Kind() != reflect.Ptr {
		return errors.New("should be pointer")
	}
	initializeVal(stype.Elem(), reflect.ValueOf(obj).Elem())
	return nil
}

// NewStruct 创建新结构体,返回是否指针和structObj保持一致
func NewStruct(structObj interface{}) interface{} {
	stype := reflect.TypeOf(structObj)
	val := newStruct(stype)
	if stype.Kind() == reflect.Ptr {
		return val.Interface()
	}
	return val.Elem().Interface()
}

const maxSliceLen = 3
const maxMapLen = 2

func randomString() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v:%v:%v", time.Now(), time.Now().Nanosecond(), rand.Float32()))))[:8]
}

func newStruct(stype reflect.Type) reflect.Value {
	if stype.Kind() == reflect.Ptr {
		stype = stype.Elem()
	}
	valPtr := reflect.New(stype)
	initializeVal(stype, valPtr.Elem())
	return valPtr
}

func initializeStruct(t reflect.Type, v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)
		var examples []string
		if ex, ok := ft.Tag.Lookup(ExampleTag); ok {
			examples = []string{ex}
		}
		if ft.Type.Kind() == reflect.Ptr {
			fv := reflect.New(ft.Type.Elem())
			initializeVal(ft.Type.Elem(), fv.Elem(), examples...)
			f.Set(fv)
		} else {
			initializeVal(ft.Type, f, examples...)
		}
	}
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

func initializeSlice(t reflect.Type, elemt reflect.Type, example ...string) reflect.Value {
	size := maxSliceLen
	if len(example) > 0 {
		list := removeString(strings.Split(example[0], ","), "")
		if len(list) > 0 {
			size = len(list)
		}
	}
	slicev := reflect.MakeSlice(t, size, size)
	if elemt.Kind() == reflect.Ptr {
		if supportExample(elemt.Elem()) && len(example) > 0 {
			list := removeString(strings.Split(example[0], ","), "")
			for i := 0; i < len(list); i++ {
				ele := reflect.New(elemt.Elem())
				initializeVal(ele.Elem().Type(), ele.Elem(), list[i])
				slicev.Index(i).Set(ele)
			}
		} else {
			for i := 0; i < size; i++ {
				ele := reflect.New(elemt.Elem())
				initializeVal(ele.Elem().Type(), ele.Elem())
				slicev.Index(i).Set(ele)
			}
		}
	} else {
		if supportExample(elemt) && len(example) > 0 {
			list := removeString(strings.Split(example[0], ","), "")
			for i := 0; i < len(list); i++ {
				ele := reflect.New(elemt)
				initializeVal(elemt, ele.Elem(), list[i])
				slicev.Index(i).Set(ele.Elem())
			}
		} else {
			for i := 0; i < size; i++ {
				ele := reflect.New(elemt)
				initializeVal(elemt, ele.Elem())
				slicev.Index(i).Set(ele.Elem())
			}
		}
	}
	return slicev
}

func supportExample(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func initializeMap(tk, tv reflect.Type, mapv reflect.Value) {
	for i := 0; i < maxMapLen; i++ {
		//key
		var key, val reflect.Value
		if tk.Kind() == reflect.Ptr {
			kptr := reflect.New(tk.Elem())
			initializeVal(kptr.Elem().Type(), kptr.Elem())
			key = kptr
		} else {
			kptr := reflect.New(tk)
			initializeVal(tk, kptr.Elem())
			key = kptr.Elem()
		}
		// value
		if tv.Kind() == reflect.Ptr {
			vptr := reflect.New(tv.Elem())
			initializeVal(vptr.Elem().Type(), vptr.Elem())
			val = vptr
		} else {
			vptr := reflect.New(tv)
			initializeVal(tv, vptr.Elem())
			val = vptr.Elem()
		}
		mapv.SetMapIndex(key, val)
	}
}

func initializeVal(t reflect.Type, v reflect.Value, examples ...string) {
	switch t.Kind() {
	case reflect.String:
		if len(examples) > 0 {
			v.SetString(examples[0])
		} else {
			v.SetString(randomString())
		}
	case reflect.Bool:
		b := rand.Intn(100)%2 == 0
		if len(examples) > 0 {
			b, _ = strconv.ParseBool(examples[0])
		}
		v.SetBool(b)
	case reflect.Int:
		if len(examples) > 0 {
			i, _ := strconv.ParseInt(examples[0], 10, 64)
			v.SetInt(i)
		} else {
			v.SetInt(rand.Int63n(10000))
		}
	case reflect.Int16:
		if len(examples) > 0 {
			i, _ := strconv.ParseInt(examples[0], 10, 64)
			v.SetInt(i)
		} else {
			v.SetInt(rand.Int63n(16))
		}
	case reflect.Int32:
		if len(examples) > 0 {
			i, _ := strconv.ParseInt(examples[0], 10, 64)
			v.SetInt(i)
		} else {
			v.SetInt(rand.Int63n(32))
		}
	case reflect.Int64:
		if len(examples) > 0 {
			i, _ := strconv.ParseInt(examples[0], 10, 64)
			v.SetInt(i)
		} else {
			v.SetInt(rand.Int63n(1000))
		}
	case reflect.Int8:
		if len(examples) > 0 {
			i, _ := strconv.ParseInt(examples[0], 10, 64)
			v.SetInt(i)
		} else {
			v.SetInt(rand.Int63n(8))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if len(examples) > 0 {
			i, _ := strconv.ParseUint(examples[0], 10, 64)
			v.SetUint(i)
		} else {
			v.SetUint(rand.Uint64() % 100)
		}
	case reflect.Float32, reflect.Float64:
		if len(examples) > 0 {
			i, _ := strconv.ParseFloat(examples[0], 64)
			v.SetFloat(i)
		} else {
			v.SetFloat(rand.Float64())
		}
	case reflect.Struct:
		if t.String() == "time.Time" {
			v.Set(reflect.ValueOf(time.Now()))
		} else {
			initializeStruct(t, v)
		}
	case reflect.Ptr:
		fv := reflect.New(t)
		initializeVal(t.Elem(), fv.Elem())
		v.Set(fv)
	case reflect.Map:
		hash := reflect.MakeMap(t)
		initializeMap(t.Key(), t.Elem(), hash)
		v.Set(hash)
	case reflect.Slice:
		array := initializeSlice(t, v.Type().Elem(), examples...)
		v.Set(array)
	case reflect.Chan:
		v.Set(reflect.MakeChan(t, 0))
	case reflect.Interface:
		v.Set(reflect.ValueOf("DYNAMIC"))
	}
}
