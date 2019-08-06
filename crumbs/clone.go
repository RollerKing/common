package crumbs

import (
	"errors"
	"reflect"
)

// Clone fromPtr to toPtr
func Clone(toPtr interface{}, fromPtr interface{}) error {
	var m uint8
	if toPtr == nil {
		m |= uint8(1)
	}
	if fromPtr == nil {
		m |= uint8(2)
	}
	// both nil
	if m == 3 {
		return nil
	}
	if m != 0 {
		return errors.New("can't clone from/to nil object")
	}
	toType := reflect.TypeOf(toPtr)
	if toType.Kind() != reflect.Ptr {
		return errors.New("to should be pointer")
	}
	fromType := reflect.TypeOf(fromPtr)
	if fromType.Kind() != reflect.Ptr {
		return errors.New("from should be pointer")
	}
	if fromType.Elem() != toType.Elem() {
		return errors.New("clone should occur in same type")
	}
	fromv, tov := reflect.ValueOf(fromPtr), reflect.ValueOf(toPtr)
	if fromv.Pointer() == tov.Pointer() {
		return errors.New("from/to is the same object")
	}
	cloneVal(fromType.Elem(), fromv.Elem(), tov.Elem())
	return nil
}

func cloneVal(t reflect.Type, fromv, tov reflect.Value) {
	switch t.Kind() {
	case reflect.String:
		tov.SetString(fromv.String())
	case reflect.Bool:
		tov.SetBool(fromv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		tov.SetInt(fromv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		tov.SetUint(fromv.Uint())
	case reflect.Float32, reflect.Float64:
		tov.SetFloat(fromv.Float())
	case reflect.Struct:
		cloneStruct(t, fromv, tov)
	case reflect.Ptr:
		if !fromv.IsNil() {
			fv := reflect.New(t)
			cloneVal(t.Elem(), fromv.Elem(), fv.Elem())
			tov.Set(fv)
		}
	case reflect.Map:
		if !fromv.IsNil() {
			hash := reflect.MakeMap(t)
			cloneMap(t.Key(), t.Elem(), fromv, hash)
			tov.Set(hash)
		}
	case reflect.Slice:
		if !fromv.IsNil() {
			slicev := reflect.MakeSlice(fromv.Type(), fromv.Len(), fromv.Cap())
			cloneSlice(fromv.Type().Elem(), fromv, slicev)
			tov.Set(slicev)
		}
	case reflect.Array:
		cloneSlice(fromv.Type().Elem(), fromv, tov)
	case reflect.Chan:
		// do not clone channel
	case reflect.Interface:
		if !fromv.IsNil() {
			if fromv.Elem().Kind() == reflect.Ptr {
				ev := reflect.New(fromv.Elem().Elem().Type())
				cloneVal(fromv.Elem().Elem().Type(), fromv.Elem().Elem(), ev.Elem())
				tov.Set(ev)
			} else {
				ev := reflect.New(fromv.Elem().Type())
				cloneVal(fromv.Elem().Type(), fromv.Elem(), ev.Elem())
				tov.Set(ev.Elem())
			}
		}
	}
}

func cloneStruct(t reflect.Type, fromv, tov reflect.Value) {
	for i := 0; i < fromv.NumField(); i++ {
		f := fromv.Field(i)
		ft := t.Field(i)
		if !isExported(ft.Name) {
			continue
		}
		if ft.Type.Kind() == reflect.Ptr {
			if f.IsNil() {
				continue
			}
			fv := reflect.New(ft.Type.Elem())
			cloneVal(ft.Type.Elem(), f.Elem(), fv.Elem())
			tov.Field(i).Set(fv)
		} else {
			cloneVal(ft.Type, f, tov.Field(i))
		}
	}
}

func cloneSlice(elemt reflect.Type, fromv, tov reflect.Value) {
	size := fromv.Len()
	if elemt.Kind() == reflect.Ptr {
		for i := 0; i < size; i++ {
			if !fromv.Index(i).IsNil() {
				ele := reflect.New(elemt.Elem())
				cloneVal(elemt.Elem(), fromv.Index(i).Elem(), ele.Elem())
				tov.Index(i).Set(ele)
			}
		}
	} else {
		for i := 0; i < size; i++ {
			cloneVal(elemt, fromv.Index(i), tov.Index(i))
		}
	}
}

func cloneMap(tk, tv reflect.Type, fromv, tov reflect.Value) {
	keys := fromv.MapKeys()
	for _, key := range keys {
		var ckey, cval reflect.Value
		// key
		if tk.Kind() == reflect.Ptr {
			kptr := reflect.New(tk.Elem())
			cloneVal(tk.Elem(), key.Elem(), kptr.Elem())
			ckey = kptr
		} else {
			kptr := reflect.New(tk)
			cloneVal(tk, key, kptr.Elem())
			ckey = kptr.Elem()
		}
		val := fromv.MapIndex(key)
		// value
		if tv.Kind() == reflect.Ptr {
			if val.IsNil() {
				cval = val
			} else {
				kptr := reflect.New(tv.Elem())
				cloneVal(tv.Elem(), val.Elem(), kptr.Elem())
				cval = kptr
			}
		} else {
			kptr := reflect.New(tv)
			cloneVal(tv, val, kptr.Elem())
			cval = kptr.Elem()
		}
		tov.SetMapIndex(ckey, cval)
	}
}
