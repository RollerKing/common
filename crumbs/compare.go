package crumbs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// NotEqualCallback when leftVal!=rightVal
type NotEqualCallback func(path string, leftVal interface{}, rightVal interface{})

// DeepCompare compare is l equals r in fact, cb can be nil
func DeepCompare(l interface{}, r interface{}, cb NotEqualCallback) bool {
	if cb == nil {
		cb = func(string, interface{}, interface{}) {}
	}

	lv, rv := reflect.ValueOf(l), reflect.ValueOf(r)
	lt, rt := lv.Type(), rv.Type()
	if lt.Kind() != rt.Kind() {
		cb("", l, r)
		return false
	}
	if lt.Kind() == reflect.Ptr && lv.Elem().Type() != rv.Elem().Type() {
		cb("", l, r)
		return false
	}
	if lt != rt {
		cb("", l, r)
		return false
	}
	return cmpVal(lt, lv, rv, []string{}, cb)
}

func cmpVal(t reflect.Type, lv, rv reflect.Value, steps []string, cb NotEqualCallback) bool {
	doCmp := func(condition bool) bool {
		if !condition {
			cb(buildCmpPath(steps), lv.Interface(), rv.Interface())
		}
		return condition
	}
	switch t.Kind() {
	case reflect.String:
		return doCmp(lv.String() == rv.String())
	case reflect.Bool:
		return doCmp(lv.Bool() == rv.Bool())
	case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int32, reflect.Int64:
		return doCmp(lv.Int() == rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return doCmp(lv.Uint() == rv.Uint())
	case reflect.Float32, reflect.Float64:
		return doCmp(lv.Float() == rv.Float())
	case reflect.Struct:
		return cmpStruct(t, lv, rv, steps, cb)
	case reflect.Ptr:
		if !lv.IsNil() && !lv.IsNil() {
			return cmpVal(t.Elem(), lv.Elem(), rv.Elem(), steps, cb)
		} else {
			return doCmp(lv.IsNil() == rv.IsNil())
		}
	case reflect.Map:
		if lv.Type() != rv.Type() {
			cb(strings.Join(steps, "."), lv.Type(), rv.Type())
			return false
		}
		if lv.Len() != rv.Len() {
			cb(buildCmpPath(steps), fmt.Sprintf("Len=%d", lv.Len()), fmt.Sprintf("Len=%d", rv.Len()))
			return false
		}
		if !lv.IsNil() && !lv.IsNil() {
			return cmpMap(t.Key(), t.Elem(), lv, rv, steps, cb)
		} else {
			return doCmp(lv.IsNil() == rv.IsNil())
		}
	case reflect.Slice, reflect.Array:
		if lv.Len() == rv.Len() && lv.Len() > 0 {
			return cmpSlice(t.Elem(), lv, rv, steps, cb)
		}
		if lv.Len() != rv.Len() {
			cb(buildCmpPath(steps), fmt.Sprintf("Len=%d", lv.Len()), fmt.Sprintf("Len=%d", rv.Len()))
			return false
		}
		return true
	case reflect.Chan:
		return true
	case reflect.Interface:
		if lv.IsNil() == rv.IsNil() {
			if lv.IsNil() {
				return true
			}
			if lv.Elem().Kind() == reflect.Ptr {
				return cmpVal(lv.Elem().Elem().Type(), lv.Elem().Elem(), rv.Elem().Elem(), steps, cb)
			} else {
				return cmpVal(lv.Elem().Type(), lv.Elem(), rv.Elem(), steps, cb)
			}
		} else {
			cb(buildCmpPath(steps), lv.Interface(), rv.Interface())
			return false
		}
	}
	return true
}

func cmpMap(k, v reflect.Type, lv, rv reflect.Value, steps []string, cb NotEqualCallback) bool {
	keys := lv.MapKeys()
	for _, key := range keys {
		lvv, rvv := lv.MapIndex(key), rv.MapIndex(key)
		if lvv.Type().Kind() == reflect.Ptr {
			if lvv.IsNil() != rvv.IsNil() {
				cb(buildCmpPath(append(steps, key.String())), lv.Interface(), rv.Interface())
				return false
			}
			if !lvv.IsNil() {
				if !cmpVal(lvv.Type().Elem(), lvv.Elem(), rvv.Elem(), append(steps, key.String()), cb) {
					return false
				}
			}
		} else {
			if !cmpVal(lvv.Type(), lvv, rvv, append(steps, key.String()), cb) {
				return false
			}
		}

	}
	return true
}

func cmpStruct(t reflect.Type, lv, rv reflect.Value, steps []string, cb NotEqualCallback) bool {
	for i := 0; i < lv.NumField(); i++ {
		lfv, rfv := lv.Field(i), rv.Field(i)
		ft := t.Field(i)
		if !isExported(ft.Name) {
			continue
		}
		if ft.Type.Kind() == reflect.Ptr {
			if lfv.IsNil() != rfv.IsNil() {
				cb(buildCmpPath(append(steps, ft.Name)), lfv.Interface(), rfv.Interface())
				return false
			}
			if !lfv.IsNil() {
				if !cmpVal(ft.Type.Elem(), lfv.Elem(), rfv.Elem(), append(steps, ft.Name), cb) {
					return false
				}
			}
		} else {
			if !cmpVal(ft.Type, lfv, rfv, append(steps, ft.Name), cb) {
				return false
			}
		}
	}
	return true
}

func cmpSlice(et reflect.Type, lv, rv reflect.Value, steps []string, cb NotEqualCallback) bool {
	if et.Kind() == reflect.Ptr {
		for i := 0; i < lv.Len(); i++ {
			if lv.Index(i).IsNil() != rv.Index(i).IsNil() {
				cb(buildCmpPath(steps), lv.Interface(), rv.Interface())
				return false
			}
			if !lv.Index(i).IsNil() {
				if !cmpVal(et.Elem(), lv.Index(i).Elem(), rv.Index(i).Elem(), append(steps, "["+strconv.FormatInt(int64(i), 10)+"]"), cb) {
					return false
				}
			}
		}
	} else {
		for i := 0; i < lv.Len(); i++ {
			if !cmpVal(et, lv.Index(i), rv.Index(i), append(steps, "["+strconv.FormatInt(int64(i), 10)+"]"), cb) {
				return false
			}
		}
	}
	return true
}

func buildCmpPath(steps []string) string {
	var list []string
	for _, s := range steps {
		if isIndexToken(s) {
			list = append(list, s)
		} else {
			list = append(list, ".", s)
		}
	}
	return strings.Join(list, "")
}

// is like [number]
func isIndexToken(s string) bool {
	token := []byte(s)
	if len(token) < 3 || token[0] != '[' || token[len(token)-1] != ']' {
		return false
	}
	// filt [00????]
	if token[1] == '0' && token[2] == '0' {
		return false
	}
	for i := 1; i < len(token)-1; i++ {
		if token[i] < '0' || token[i] > '9' {
			return false
		}
	}
	return true
}
