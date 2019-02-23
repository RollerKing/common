package fixture

import (
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"
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
		if ft.Type.Kind() == reflect.Ptr {
			fv := reflect.New(ft.Type.Elem())
			initializeVal(ft.Type.Elem(), fv.Elem())
			f.Set(fv)
		} else {
			initializeVal(ft.Type, f)
		}
	}
}

func initializeSlice(elemt reflect.Type, slicev reflect.Value) {
	if elemt.Kind() == reflect.Ptr {
		for i := 0; i < maxSliceLen; i++ {
			ele := reflect.New(elemt.Elem())
			initializeVal(ele.Elem().Type(), ele.Elem())
			slicev.Index(i).Set(ele)
		}
	} else {
		for i := 0; i < maxSliceLen; i++ {
			ele := reflect.New(elemt)
			initializeVal(elemt, ele.Elem())
			slicev.Index(i).Set(ele.Elem())
		}
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

func initializeVal(t reflect.Type, v reflect.Value) {
	switch t.Kind() {
	case reflect.String:
		v.SetString(randomString())
	case reflect.Bool:
		b := rand.Intn(100)%2 == 0
		v.SetBool(b)
	case reflect.Int:
		v.SetInt(rand.Int63n(10000))
	case reflect.Int16:
		v.SetInt(rand.Int63n(16))
	case reflect.Int32:
		v.SetInt(rand.Int63n(32))
	case reflect.Int64:
		v.SetInt(rand.Int63n(1000))
	case reflect.Int8:
		v.SetInt(rand.Int63n(8))
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
		array := reflect.MakeSlice(t, maxSliceLen, maxSliceLen)
		initializeSlice(v.Type().Elem(), array)
		v.Set(array)
	case reflect.Chan:
		v.Set(reflect.MakeChan(t, 0))
	}
}
