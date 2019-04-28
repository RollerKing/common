package crumbs

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// IndexStructs 将结构体数组按某个字段名去重成为map
func IndexStructs(distMap interface{}, srcArray interface{}, field string) {
	src := reflect.ValueOf(srcArray)
	dest := reflect.ValueOf(distMap)
	if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
		panic(fmt.Sprintf("src[%v] must be array", src.Kind()))
	}
	// validate map
	if !dest.IsValid() {
		panic("dest map invalid")
	}
	if dest.IsNil() {
		panic("dest map is nil")
	}
	if dest.Kind() != reflect.Map {
		panic(fmt.Sprintf("dest[%v] must be map", dest.Kind()))
	}
	isSrcPtr := false
	srcType := src.Type().Elem()
	if src.Type().Elem().Kind() == reflect.Ptr {
		if src.Type().Elem().Elem().Kind() != reflect.Struct {
			panic(fmt.Sprintf("src must be struct array:%v", src.Type().Elem().Kind()))
		}
		isSrcPtr = true
		srcType = src.Type().Elem().Elem()
	} else if src.Type().Elem().Kind() == reflect.Struct {
		//ok
	} else {
		panic(fmt.Sprintf("src must be struct array:%v", src.Type().Elem().Kind()))
	}
	totalFields := srcType.NumField()
	if totalFields == 0 {
		return
	}
	iField := -1
	for i := 0; i < totalFields; i++ {
		if srcType.Field(i).Name == field {
			iField = i
			break
		}
	}
	if iField < 0 {
		panic("no such field " + field)
	}
	if dest.Type().Key().Kind() != srcType.Field(iField).Type.Kind() {
		panic(fmt.Sprintf("key[%v] of dest must be same with field type", dest.Type().Key().Kind()))
	}
	if src.Type().Elem().Kind() != dest.Type().Elem().Kind() {
		panic(fmt.Sprintf("dest[%v] and src[%v] element type should be same", dest.Type().Elem().Kind(), src.Type().Elem().Kind()))
	}
	if src.Type().Elem().String() != dest.Type().Elem().String() {
		panic(fmt.Sprintf("dest[%v] and src[%v] element type should be same", dest.Type().Elem().String(), src.Type().Elem().String()))
	}

	length := src.Len()
	for i := 0; i < length; i++ {
		if isSrcPtr {
			dest.SetMapIndex(src.Index(i).Elem().FieldByName(field), src.Index(i))
		} else {
			dest.SetMapIndex(src.Index(i).FieldByName(field), src.Index(i))
		}
	}
}

// BucketStructs 将结构体数组按某个字段名归并为数组的map
func BucketStructs(distMap interface{}, srcArray interface{}, field string) {
	src := reflect.ValueOf(srcArray)
	dest := reflect.ValueOf(distMap)
	if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
		panic(fmt.Sprintf("src[%v] must be array", src.Kind()))
	}
	// validate map
	if !dest.IsValid() {
		panic("dest map invalid")
	}
	if dest.IsNil() {
		panic("dest map is nil")
	}
	if dest.Kind() != reflect.Map {
		panic(fmt.Sprintf("dest[%v] must be map", dest.Kind()))
	}
	isSrcPtr := false
	srcType := src.Type().Elem()
	if src.Type().Elem().Kind() == reflect.Ptr {
		if src.Type().Elem().Elem().Kind() != reflect.Struct {
			panic(fmt.Sprintf("src must be struct array:%v", src.Type().Elem().Kind()))
		}
		isSrcPtr = true
		srcType = src.Type().Elem().Elem()
	} else if src.Type().Elem().Kind() == reflect.Struct {
		//ok
	} else {
		panic(fmt.Sprintf("src must be struct array:%v", src.Type().Elem().Kind()))
	}
	totalFields := srcType.NumField()
	if totalFields == 0 {
		return
	}
	iField := -1
	for i := 0; i < totalFields; i++ {
		if srcType.Field(i).Name == field {
			iField = i
			break
		}
	}
	if iField < 0 {
		panic("no such field " + field)
	}
	if dest.Type().Key().Kind() != srcType.Field(iField).Type.Kind() {
		panic(fmt.Sprintf("key[%v] of dest must be same with field type", dest.Type().Key().Kind()))
	}
	if dest.Type().Elem().Kind() != reflect.Slice {
		panic("dest value should be slice")
	}
	if src.Type().Elem().Kind() != dest.Type().Elem().Elem().Kind() {
		panic(fmt.Sprintf("dest[%v] and src[%v] element type should be same", dest.Type().Elem().Kind(), src.Type().Elem().Kind()))
	}
	if src.Type().Elem().String() != dest.Type().Elem().Elem().String() {
		panic(fmt.Sprintf("dest[%v] and src[%v] element type should be same", dest.Type().Elem().Elem().String(), src.Type().Elem().String()))
	}

	length := src.Len()
	srcSliceType := reflect.TypeOf(srcArray)
	for i := 0; i < length; i++ {
		var key reflect.Value
		if isSrcPtr {
			key = src.Index(i).Elem().FieldByName(field)
		} else {
			key = src.Index(i).FieldByName(field)
		}
		mVal := dest.MapIndex(key)
		if !mVal.IsValid() || mVal.IsNil() {
			mVal = reflect.MakeSlice(srcSliceType, 0, 0)
		}
		mVal = reflect.Append(mVal, src.Index(i))
		dest.SetMapIndex(key, mVal)
	}
}

// SortStructs 将结构体数组按某个字段排序 e.g. SortStructs([]Goods{g1,g2},"GoodsId")
func SortStructs(srcArray interface{}, field string) error {
	if srcArray == nil {
		return nil
	}
	src := reflect.ValueOf(srcArray)
	if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
		return fmt.Errorf("src[%v] must be array", src.Kind())
	}
	length := src.Len()
	if length <= 1 {
		return nil
	}
	isSrcPtr := false
	srcType := src.Type().Elem()
	if src.Type().Elem().Kind() == reflect.Ptr {
		if src.Type().Elem().Elem().Kind() != reflect.Struct {
			return fmt.Errorf("src must be struct array:%v", src.Type().Elem().Kind())
		}
		isSrcPtr = true
		srcType = src.Type().Elem().Elem()
	} else if src.Type().Elem().Kind() == reflect.Struct {
		//ok
	} else {
		return fmt.Errorf("src must be struct array:%v", src.Type().Elem().Kind())
	}
	iField := -1
	for i := 0; i < srcType.NumField(); i++ {
		if srcType.Field(i).Name == field {
			iField = i
			break
		}
	}
	if iField < 0 {
		return fmt.Errorf("no such field %s", field)
	}

	ss := &structSlice{
		data: make([]structWrapper, length),
	}
	if isSrcPtr {
		ss.tp = src.Index(0).Elem().Field(iField).Type()
	} else {
		ss.tp = src.Index(0).Field(iField).Type()
	}
	for i := 0; i < length; i++ {
		ss.data[i].index = i
		if isSrcPtr {
			ss.data[i].field = src.Index(i).Elem().Field(iField)
		} else {
			ss.data[i].field = src.Index(i).Field(iField)
		}
	}
	sort.Sort(ss)
	checked := make(map[int]int)
	basei, to := 0, 0
	from := ss.data[to].index
	base := reflect.New(srcType)
	if !isSrcPtr {
		base = base.Elem()
		base.Set(src.Index(to))
	} else {
		base.Elem().Set(src.Index(to).Elem())
	}
	for cnt := 0; cnt < length; cnt++ {
		checked[to] = 1
		if from == basei || from == to {
			if !isSrcPtr {
				src.Index(to).Set(base)
			} else {
				src.Index(to).Elem().Set(base.Elem())
			}
			for i := 0; i < length; i++ {
				if _, ok := checked[i]; !ok {
					basei = i
					base = reflect.New(srcType)
					if !isSrcPtr {
						base = base.Elem()
						base.Set(src.Index(i))
					} else {
						base.Elem().Set(src.Index(i).Elem())
					}
					from, to = ss.data[basei].index, basei
					break
				}
			}
			continue
		}
		if !isSrcPtr {
			src.Index(to).Set(src.Index(from))
		} else {
			src.Index(to).Elem().Set(src.Index(from).Elem())
		}
		to = from
		from = ss.data[from].index
	}
	return nil
}

var typeOfTime = reflect.TypeOf(time.Time{})

//Be careful to use, from,to must be pointer
func DumpStruct(to interface{}, from interface{}) {
	fromv := reflect.ValueOf(from)
	tov := reflect.ValueOf(to)
	if fromv.Kind() != reflect.Ptr || tov.Kind() != reflect.Ptr {
		return
	}

	from_val := reflect.Indirect(fromv)
	to_val := reflect.Indirect(tov)

	for i := 0; i < from_val.Type().NumField(); i++ {
		fdi_from_val := from_val.Field(i)
		fd_name := from_val.Type().Field(i).Name
		fdi_to_val := to_val.FieldByName(fd_name)

		if !fdi_to_val.IsValid() || fdi_to_val.Kind() != fdi_from_val.Kind() {
			continue
		}

		switch fdi_from_val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if fdi_to_val.Type() != fdi_from_val.Type() {
				fdi_to_val.Set(fdi_from_val.Convert(fdi_to_val.Type()))
			} else {
				fdi_to_val.Set(fdi_from_val)
			}
		case reflect.Slice:
			if fdi_to_val.IsNil() {
				fdi_to_val.Set(reflect.MakeSlice(fdi_to_val.Type(), fdi_from_val.Len(), fdi_from_val.Len()))
			}
			DumpList(fdi_to_val.Interface(), fdi_from_val.Interface())
		case reflect.Struct:
			if fdi_to_val.Type() == typeOfTime {
				if fdi_to_val.Type() != fdi_from_val.Type() {
					continue
				}
				fdi_to_val.Set(fdi_from_val)
			} else {
				DumpStruct(fdi_to_val.Addr(), fdi_from_val.Addr())
			}
		default:
			if fdi_to_val.Type() != fdi_from_val.Type() {
				continue
			}
			fdi_to_val.Set(fdi_from_val)
		}
	}
}

//Be careful to use, from,to must be pointer
func DumpList(to interface{}, from interface{}) {
	raw_to := reflect.ValueOf(to)
	//raw_from := reflect.ValueOf(from)

	val_from := reflect.Indirect(reflect.ValueOf(from))
	val_to := reflect.Indirect(reflect.ValueOf(to))

	if !(val_from.Kind() == reflect.Slice) || !(val_to.Kind() == reflect.Slice) {
		return
	}

	if raw_to.Kind() == reflect.Ptr && raw_to.Elem().Len() == 0 {
		val_to.Set(reflect.MakeSlice(val_to.Type(), val_from.Len(), val_from.Len()))
	}

	if val_from.Len() == val_to.Len() {
		for i := 0; i < val_from.Len(); i++ {
			switch val_from.Index(i).Kind() {
			case reflect.Struct:
				DumpStruct(val_to.Index(i).Addr().Interface(), val_from.Index(i).Addr().Interface())
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.String:
				val_to.Index(i).Set(val_from.Index(i))
			default:
				continue
			}
		}
	}
}

func isExported(fieldName string) bool {
	return len(fieldName) > 0 && (fieldName[0] >= 'A' && fieldName[0] <= 'Z')
}

// Struct2Map convert struct to map[string]interface{}
func Struct2Map(obj interface{}, tagName ...string) map[string]interface{} {
	res := make(map[string]interface{})
	val := reflect.ValueOf(obj)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	getKeyFunc := func(f reflect.StructField) (name string, omitempty bool, drop bool) {
		for _, tag := range tagName {
			if tag != "" {
				if arr := strings.SplitN(f.Tag.Get(tag), ",", 2); len(arr) > 0 && arr[0] != "" {
					name = arr[0]
					omitempty = strings.Contains(f.Tag.Get(tag), ",omitempty")
					drop = name == "-"
					return
				}
			}
		}
		return f.Name, false, false
	}
	for i := 0; i < val.Type().NumField(); i++ {
		name, omitempty, dropField := getKeyFunc(val.Type().Field(i))
		if dropField {
			continue
		}
		field := val.Field(i)
		kind := field.Kind()
		isPtr := field.Type().Kind() == reflect.Ptr
		if !isExported(val.Type().Field(i).Name) {
			continue
		}
		if isPtr {
			if field.IsNil() {
				if !omitempty {
					res[name] = nil
				}
				continue
			}
			field = field.Elem()
			kind = field.Kind()
		}
		if omitempty && isZeroValue(field.Type(), field, false) {
			continue
		}
		switch kind {
		case reflect.Slice:
			list := make([]interface{}, field.Len())
			for j := 0; j < field.Len(); j++ {
				iv := field.Index(j)
				if iv.Type().Kind() == reflect.Ptr {
					iv = iv.Elem()
				}
				if iv.Kind() == reflect.Struct && iv.Type().String() != "time.Time" {
					list[j] = Struct2Map(iv.Interface(), tagName...)
				} else {
					list[j] = iv.Interface()
				}
			}
			res[name] = list
		case reflect.Struct:
			if field.Type().String() == "time.Time" {
				res[name] = field.Interface()
			} else {
				sub := Struct2Map(field.Interface(), tagName...)
				if val.Type().Field(i).Anonymous {
					for k, v := range sub {
						res[k] = v
					}
				} else {
					res[name] = sub
				}
			}
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.String:
			res[name] = field.Interface()
		default:
			res[name] = field.Interface()
		}
	}
	return res
}

// IsZeroStruct is empty struct
func IsZeroStruct(st interface{}) bool {
	tp := reflect.TypeOf(st)
	val := reflect.ValueOf(st)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
		val = val.Elem()
	}
	return isZeroValue(tp, val, false)
}

// IsStrictZeroStruct is empty struct, would check ptr content
func IsStrictZeroStruct(st interface{}) bool {
	tp := reflect.TypeOf(st)
	val := reflect.ValueOf(st)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
		val = val.Elem()
	}
	return isZeroValue(tp, val, true)
}

func isBaseZeroValue(tp reflect.Type, val reflect.Value) (isbase bool, isempty bool) {
	switch tp.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true, reflect.Zero(tp).Int() == val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true, reflect.Zero(tp).Uint() == val.Uint()
	case reflect.Bool:
		return true, !reflect.Zero(tp).Bool()
	case reflect.Float32, reflect.Float64:
		return true, reflect.Zero(tp).Float() == val.Float()
	case reflect.String:
		return true, reflect.Zero(tp).String() == val.String()
	}
	return false, false
}

func isZeroValue(tp reflect.Type, val reflect.Value, checkPtrContent bool) bool {
	if tp.Kind() == reflect.Ptr {
		if checkPtrContent {
			if val.IsNil() {
				return true
			}
			tp = tp.Elem()
			val = val.Elem()
		} else {
			return val.IsNil()
		}
	}
	switch tp.Kind() {
	case reflect.Map, reflect.Array, reflect.Slice, reflect.Chan:
		return val.Len() == 0
	case reflect.Func:
		return val.IsNil()
	case reflect.Struct:
		for i := 0; i < tp.NumField(); i++ {
			ft := tp.Field(i)
			f := val.Field(i)
			if empty := isZeroValue(ft.Type, f, checkPtrContent); !empty {
				return false
			}
		}
		return true
	default:
		if isbase, isempty := isBaseZeroValue(tp, val); isbase {
			return isempty
		}
		if !val.CanInterface() {
			return true
		}
		return val.Interface() == reflect.Zero(tp).Interface()
	}
}

type structWrapper struct {
	field reflect.Value
	index int
}

type structSlice struct {
	data []structWrapper
	tp   reflect.Type
}

func (s *structSlice) Len() int {
	return len(s.data)
}
func (s *structSlice) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *structSlice) Less(i, j int) bool {
	switch s.tp.Kind() {
	case reflect.String:
		return s.data[i].field.String() < s.data[j].field.String()
	case reflect.Float32, reflect.Float64:
		return s.data[i].field.Float() < s.data[j].field.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return s.data[i].field.Int() < s.data[j].field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return s.data[i].field.Uint() < s.data[j].field.Uint()
	case reflect.Struct:
		if s.tp == typeOfTime {
			return s.data[i].field.Interface().(time.Time).Before(s.data[j].field.Interface().(time.Time))
		}
	}
	return fmt.Sprint(s.data[i].field.Interface()) < fmt.Sprint(s.data[j].field.Interface())
}
