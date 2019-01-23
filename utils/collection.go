package utils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// 求交集
func InteractStrings(lists ...[]string) []string {
	for i := range lists {
		if len(lists[i]) == 0 {
			return nil
		}
		sort.Strings(lists[i])
	}
	return interactList(lists, 0, len(lists)-1)
}

func interactStrings(list1 []string, list2 []string) []string {
	if len(list1) == 0 || len(list2) == 0 {
		return nil
	}
	var res []string
	for i, j := 0, 0; i < len(list1) && j < len(list2); {
		if list1[i] == list2[j] {
			res = append(res, list1[i])
			i++
			j++
		} else if list1[i] > list2[j] {
			j++
		} else {
			i++
		}
	}
	return res
}

func interactList(lists [][]string, start, end int) []string {
	if start > end {
		return nil
	}
	mid := (start + end) / 2
	var left, right []string
	switch mid - start {
	case 0:
		left = lists[start]
	case 1:
		left = interactStrings(lists[start], lists[mid])
	default:
		left = interactList(lists, start, mid)
	}
	mid++
	if mid <= end {
		switch end - mid {
		case 0:
			right = lists[mid]
		case 1:
			right = interactStrings(lists[mid], lists[end])
		default:
			right = interactList(lists, mid, end)
		}
	}
	return interactStrings(left, right)
}

// 求差集
func SubstractStrings(list1 []string, list2 []string) []string {
	sort.Strings(list1)
	sort.Strings(list2)
	var res []string
	var i, j int
	for i < len(list1) && j < len(list2) {
		if list1[i] < list2[j] {
			res = append(res, list1[i])
			i++
		} else if list1[i] == list2[j] {
			i++
			j++
		} else {
			j++
		}
	}
	if i < len(list1) {
		res = append(res, list1[i:]...)
	}
	return res
}

// 求并集
func MergeStrings(lists ...[]string) []string {
	for i := range lists {
		sort.Strings(lists[i])
	}
	return andList(lists, 0, len(lists)-1)
}

func andList(lists [][]string, start, end int) []string {
	if start > end {
		return nil
	}
	mid := (start + end) / 2
	var left, right []string
	switch mid - start {
	case 0:
		left = lists[start]
	case 1:
		left = andStrings(lists[start], lists[mid])
	default:
		left = andList(lists, start, mid)
	}
	mid++
	if mid <= end {
		switch end - mid {
		case 0:
			right = lists[mid]
		case 1:
			right = andStrings(lists[mid], lists[end])
		default:
			right = andList(lists, mid, end)
		}
	}
	return andStrings(left, right)
}

func andStrings(list1 []string, list2 []string) []string {
	var res []string
	var i, j int
	for i < len(list1) && j < len(list2) {
		if list1[i] == list2[j] {
			res = append(res, list1[i])
			i++
			j++
		} else if list1[i] > list2[j] {
			res = append(res, list2[j])
			j++
		} else {
			res = append(res, list1[i])
			i++
		}
	}
	if i < len(list1) {
		res = append(res, list1[i:]...)
	}
	if j < len(list2) {
		res = append(res, list2[j:]...)
	}
	return res
}

func EqualStrings(list1, list2 []string) bool {
	if len(list1) != len(list2) {
		return false
	}
	sort.Strings(list1)
	sort.Strings(list2)
	for i := 0; i < len(list1); i++ {
		if list1[i] != list2[i] {
			return false
		}
	}
	return true
}

func UniqStrings(list []string) []string {
	memo := make(map[string]int)
	for _, e := range list {
		memo[e] = 1
	}
	var arr []string
	for str := range memo {
		arr = append(arr, str)
	}
	return arr
}

func ContainString(list []string, target string) bool {
	for _, e := range list {
		if e == target {
			return true
		}
	}
	return false
}

func ContainStringIgnoreCase(list []string, target string) bool {
	target = strings.ToLower(target)
	for _, e := range list {
		if strings.ToLower(e) == target {
			return true
		}
	}
	return false
}

func PartitionStrings(list []string, size int) [][]string {
	var res [][]string
	for start := 0; start < len(list); start += size {
		end := start + size
		if end > len(list) {
			end = len(list)
		}
		res = append(res, list[start:end])
	}
	return res
}

func MapStrings(list []string, f func(string) string) []string {
	for i, str := range list {
		list[i] = f(str)
	}
	return list
}

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
