package aux

import (
	"errors"
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

// SortStructs 将结构体数组按某个字段排序
// e.g. SortStructs([]string{"aa","bb"},[]Goods{g1,g2},"GoodsId")
func SortStructs(indexes interface{}, srcArray interface{}, field string) {
	orders := make(map[string][]int)
	idxVal := reflect.ValueOf(indexes)
	if idxVal.Kind() != reflect.Slice && idxVal.Kind() != reflect.Array {
		panic("indexes must be array")
	}
	idxLength := idxVal.Len()
	for i := 0; i < idxLength; i++ {
		key := fmt.Sprint(idxVal.Index(i))
		if arr, ok := orders[key]; ok {
			orders[key] = append(arr, i)
		} else {
			orders[key] = []int{i}
		}
	}
	src := reflect.ValueOf(srcArray)
	if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
		panic(fmt.Sprintf("src[%v] must be array", src.Kind()))
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

	length := src.Len()
	if length != idxLength {
		panic("index and values array length not match")
	}
	copyArray := reflect.MakeSlice(reflect.TypeOf(srcArray), length, length)
	reflect.Copy(copyArray, src)
	for i := 0; i < length; i++ {
		var key string
		if isSrcPtr {
			key = fmt.Sprint(copyArray.Index(i).Elem().FieldByName(field))
		} else {
			key = fmt.Sprint(copyArray.Index(i).FieldByName(field))
		}
		j := 0
		jrr := orders[key]
		if len(jrr) > 0 {
			j = jrr[0]
			orders[key] = jrr[1:]
		}
		src.Index(j).Set(copyArray.Index(i))
	}
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

func Len(in interface{}) int {
	v := reflect.Indirect(reflect.ValueOf(in))
	return v.Len()

}

type structslice struct {
	Key string
	Val reflect.Value
}

func (a structslice) Len() int { return a.Val.Len() }
func (a structslice) Less(i, j int) bool {
	vi := a.Val.Index(i).FieldByName(a.Key)
	vj := a.Val.Index(j).FieldByName(a.Key)
	switch vi.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return vi.Int() < vj.Int()
	case reflect.String:
		return vi.String() < vj.String()
	}
	return i < j
}
func (a structslice) Swap(i, j int) {
	x, y := a.Val.Index(i).Interface(), a.Val.Index(j).Interface()
	a.Val.Index(i).Set(reflect.ValueOf(y))
	a.Val.Index(j).Set(reflect.ValueOf(x))
}

type slice struct {
	structslice
	m map[interface{}]int
}

func (a slice) Less(i, j int) bool {
	vi := a.Val.Index(i).FieldByName(a.Key)
	vj := a.Val.Index(j).FieldByName(a.Key)
	return a.m[vi.Interface()] < a.m[vj.Interface()]
}

func (a *slice) Init(list, by reflect.Value, key string) {
	a.Key = key
	a.Val = list
	a.m = make(map[interface{}]int)
	for i := 0; i < by.Len(); i++ {
		a.m[by.Index(i).Interface()] = i
	}
}

//sort slice of structs by the given order of the field
func SortList(list, by interface{}, key string) {
	val_list := reflect.ValueOf(list)
	val_by := reflect.ValueOf(by)

	if val_list.Type().Kind() == reflect.Slice && val_by.Type().Kind() == reflect.Slice &&
		val_list.Len() == val_by.Len() {
		a := slice{}
		a.Init(val_list, val_by, key)
		sort.Sort(a)
	}
}

//in must be slice
//f must be func with one input parameter and one output parameter
//type of f's input parameter must be the same type as in's elem
func Map(in interface{}, f interface{}) interface{} {
	valIn := reflect.ValueOf(in)
	valF := reflect.ValueOf(f)
	if valIn.Kind() != reflect.Slice {
		return nil
	}
	if valF.Kind() != reflect.Func {
		return nil
	}
	tyF := valF.Type()
	if tyF.NumIn() != 1 || tyF.NumOut() != 1 {
		return nil
	}
	if tyF.In(0) != valIn.Type().Elem() {
		return nil
	}

	l := valIn.Len()
	valRs := reflect.MakeSlice(reflect.SliceOf(tyF.Out(0)), 0, l)
	for i := 0; i < l; i++ {
		out := valF.Call([]reflect.Value{valIn.Index(i)})
		valRs = reflect.Append(valRs, out[0])
	}
	if valRs.IsValid() {
		return valRs.Interface()
	}
	return nil
}

//in must be slice
//f must be func with one input parameter and two output parameter
//type of f's input parameter must be the same type as in's elem
//type of f's 2nd output parameter must be Bool
func MapFilter(in interface{}, f interface{}) interface{} {
	valIn := reflect.ValueOf(in)
	valF := reflect.ValueOf(f)
	if valIn.Kind() != reflect.Slice {
		return nil
	}
	if valF.Kind() != reflect.Func {
		return nil
	}
	tyF := valF.Type()
	if tyF.NumIn() != 1 || tyF.NumOut() != 2 {
		return nil
	}
	if tyF.In(0) != valIn.Type().Elem() {
		return nil
	}
	if reflect.Bool != tyF.Out(1).Kind() {
		return nil
	}

	l := valIn.Len()
	valRs := reflect.MakeSlice(reflect.SliceOf(tyF.Out(0)), 0, 0)
	for i := 0; i < l; i++ {
		out := valF.Call([]reflect.Value{valIn.Index(i)})
		if out[1].Bool() {
			valRs = reflect.Append(valRs, out[0])
		}
	}
	return valRs.Interface()
}

//in must be slice
//f must be func with one input parameter and one output parameter
//type of f's input parameter must be the same type as in's elem
//type of f's output parameter must be Bool
func Filter(in interface{}, f interface{}) interface{} {
	valIn := reflect.ValueOf(in)
	valF := reflect.ValueOf(f)
	if valIn.Kind() != reflect.Slice {
		return nil
	}
	if valF.Kind() != reflect.Func {
		return nil
	}
	tyF := valF.Type()
	if tyF.NumIn() != 1 || tyF.NumOut() != 1 {
		return nil
	}
	if tyF.In(0) != valIn.Type().Elem() {
		return nil
	}
	if reflect.Bool != tyF.Out(0).Kind() {
		return nil
	}

	valRs := reflect.MakeSlice(valIn.Type(), 0, 0)
	for i := 0; i < valIn.Len(); i++ {
		out := valF.Call([]reflect.Value{valIn.Index(i)})
		if out[0].Bool() {
			valRs = reflect.Append(valRs, valIn.Index(i))
		}
	}
	return valRs.Interface()
}

//in must be pointer to slice
func InitSturctSlice(in interface{}, extra ...interface{}) {
	valIn := reflect.ValueOf(in)
	if valIn.Kind() != reflect.Ptr {
		return
	}
	valIn = reflect.Indirect(valIn)
	if valIn.Kind() != reflect.Slice {
		return
	}

	l, c := 0, 0
	if len(extra) > 0 {
		if d, ok := extra[0].(int); ok {
			l, c = d, d
		}
	}
	if len(extra) > 1 {
		if d, ok := extra[1].(int); ok {
			if d > l {
				c = d
			}
		}
	}

	valRs := reflect.MakeSlice(valIn.Type(), l, c)
	valIn.Set(valRs)
}

func LeftJoin(left, right interface{}, leftOn, rightOn string, onJoin interface{}) {
	valL := reflect.ValueOf(left)
	valR := reflect.ValueOf(right)
	if valL.Kind() != reflect.Slice || valR.Kind() != reflect.Slice {
		return
	}

	tyL := valL.Type().Elem()
	tyR := valR.Type().Elem()
	if tyL.Kind() != reflect.Struct || tyR.Kind() != reflect.Struct {
		return
	}

	valJoin := reflect.ValueOf(onJoin)
	tyJoin := reflect.TypeOf(onJoin)
	if tyJoin.Kind() != reflect.Func {
		return
	}
	if tyJoin.NumIn() != 2 {
		return
	}
	if tyJoin.In(0) != reflect.PtrTo(tyL) || tyJoin.In(1) != reflect.PtrTo(tyR) {
		return
	}

	fieldL, ok := valL.Type().Elem().FieldByName(leftOn)
	if !ok {
		return
	}
	fieldR, ok := valR.Type().Elem().FieldByName(rightOn)
	if !ok {
		return
	}
	if fieldL.Type != fieldR.Type {
		return
	}

	m := reflect.MakeMap(reflect.MapOf(fieldR.Type, reflect.PtrTo(tyR)))

	for i := 0; i < valR.Len(); i++ {
		val := valR.Index(i).Addr()
		key := valR.Index(i).FieldByName(rightOn)
		m.SetMapIndex(key, val)
	}

	for i := 0; i < valL.Len(); i++ {
		vL := valL.Index(i).Addr()
		key := valL.Index(i).FieldByName(leftOn)
		vR := m.MapIndex(key)
		if !vR.IsValid() {
			vR = reflect.Zero(reflect.PtrTo(tyR))
		}
		valJoin.Call([]reflect.Value{vL, vR})
	}
}

/*
 * 参数默认值
 */
const (
	IntDefaultValue   = -999999999
	CharDefaultValue  = "Codoon_2016_02_24_10_47_25"
	FloatDefaultValue = -999999999.0
)

//Set default value for each struct field by field type.
//Each specify type value is aboved.
func InitBySpecifyValue(data interface{}) error {
	reflect_value := reflect.ValueOf(data).Elem()
	if !reflect_value.CanSet() {
		return errors.New("Data can not set.")
	}
	for i := 0; i < reflect_value.NumField(); i++ {
		switch reflect_value.Field(i).Type().Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			reflect_value.Field(i).SetInt(IntDefaultValue)
		case reflect.String:
			reflect_value.Field(i).SetString(CharDefaultValue)
		case reflect.Float32, reflect.Float64:
			reflect_value.Field(i).SetFloat(FloatDefaultValue)
		}
	}

	return nil
}

func IsDefaultIntValue(value interface{}) bool {
	return reflect.ValueOf(value).Int() == IntDefaultValue
}

func IsDefaultCharValue(value string) bool {
	return value == CharDefaultValue
}

func IsDefaultFloatValue(value interface{}) bool {
	return reflect.ValueOf(value).Float() == FloatDefaultValue
}

func Dedup(slice interface{}) {
	vPtr := reflect.ValueOf(slice)
	if vPtr.Kind() != reflect.Ptr {
		return
	}
	v := vPtr.Elem()
	if v.Kind() != reflect.Slice {
		return
	}
	if v.Len() < 2 {
		return
	}

	uniq := reflect.MakeSlice(v.Type(), 0, v.Len())
	tMap := reflect.MapOf(v.Type().Elem(), reflect.TypeOf(true))

	m := reflect.MakeMap(tMap)

	vOk := reflect.ValueOf(true)
	for i := 0; i < v.Len(); i++ {
		vi := v.Index(i)
		if ok := m.MapIndex(vi); !ok.IsValid() {
			uniq = reflect.Append(uniq, vi)
			m.SetMapIndex(vi, vOk)
		}
	}
	v.Set(uniq)
}

// Struct2Map convert struct to map[string]interface{}
func Struct2Map(obj interface{}, tagName ...string) map[string]interface{} {
	res := make(map[string]interface{})
	val := reflect.ValueOf(obj)
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	getKeyFunc := func(f reflect.StructField) (name string, omitempty bool) {
		for _, tag := range tagName {
			if tag != "" {
				if arr := strings.SplitN(f.Tag.Get(tag), ",", 2); len(arr) > 0 && arr[0] != "" {
					name = arr[0]
					omitempty = strings.Contains(f.Tag.Get(tag), ",omitempty")
					return
				}
			}
		}
		return f.Name, false
	}
	for i := 0; i < val.Type().NumField(); i++ {
		name, omitempty := getKeyFunc(val.Type().Field(i))
		field := val.Field(i)
		kind := field.Kind()
		isPtr := field.Type().Kind() == reflect.Ptr
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
		switch kind {
		case reflect.Slice:
			list := make([]interface{}, field.Len())
			for j := 0; j < field.Len(); j++ {
				iv := field.Index(j)
				if iv.Type().Kind() == reflect.Ptr {
					iv = iv.Elem()
				}
				if iv.Kind() == reflect.Struct && iv.Type().String() != "time.Time" {
					list[j] = Struct2Map(iv.Interface())
				} else {
					list[j] = iv.Interface()
				}
			}
			res[name] = list
		case reflect.Struct:
			if field.Type().String() == "time.Time" {
				res[name] = field.Interface()
			} else {
				res[name] = Struct2Map(field.Interface())
			}
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.String:
			res[name] = field.Interface()
		default:
			res[name] = field.Interface()
		}
	}
	return res
}
