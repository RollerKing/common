package crumbs

import (
	"sort"
)

// 包含
func ContainInt(nums []int, num int) bool {
	for _, n := range nums {
		if n == num {
			return true
		}
	}
	return false
}

// 求并集
func UnionInts(lists ...[]int) []int {
	for i := range lists {
		sort.Ints(lists[i])
	}
	return andIntList(lists, 0, len(lists)-1)
}

func andIntList(lists [][]int, start, end int) []int {
	if start > end {
		return nil
	}
	mid := (start + end) / 2
	var left, right []int
	switch mid - start {
	case 0:
		left = lists[start]
	case 1:
		left = andInts(lists[start], lists[mid])
	default:
		left = andIntList(lists, start, mid)
	}
	mid++
	if mid <= end {
		switch end - mid {
		case 0:
			right = lists[mid]
		case 1:
			right = andInts(lists[mid], lists[end])
		default:
			right = andIntList(lists, mid, end)
		}
	}
	return andInts(left, right)
}

func andInts(list1 []int, list2 []int) []int {
	var res []int
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
