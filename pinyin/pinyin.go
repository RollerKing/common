package pinyin

import (
	gpy "github.com/mozillazg/go-pinyin"
	"strings"
	"unicode/utf8"
)

// FuzzyContains s是否包含substr
func FuzzyContains(s string, substr string) bool {
	if strings.Contains(s, substr) {
		return true
	}
	idx, _ := PinyinContains(s, substr)
	return idx >= 0
}

// FuzzyIndex substr在s中位置,并返回匹配到的原汉字串
func FuzzyIndex(s string, substr string) (int, string) {
	if idx := strings.Index(s, substr); idx >= 0 {
		return idx, substr
	}
	idx, size := PinyinContains(s, substr)
	if idx >= 0 {
		return idx, s[idx : idx+size]
	}
	return -1, ""
}

// ToPinyinString 汉字->拼音
func ToPinyinString(hanzi string) string {
	s, _ := ToPinyin(hanzi)
	return s
}

// ToPinyinStringWithSpace 汉字->拼音,转化后汉字用空格分割
func ToPinyinStringWithSpace(hanzi string) string {
	args := gpy.NewArgs()
	var text []string
	prevIsHanzi := true
	for _, r := range hanzi {
		py := gpy.SinglePinyin(r, args)
		if len(py) > 0 {
			text = append(text, py[0])
			prevIsHanzi = true
		} else {
			if prevIsHanzi {
				text = append(text, string(r))
			} else {
				text[len(text)-1] = text[len(text)-1] + string(r)
			}
			prevIsHanzi = false
		}
	}
	return strings.Join(text, " ")
}

// ToPinyin 汉字->拼音并标记位置
func ToPinyin(raw string) (string, [][3]int) {
	args := gpy.NewArgs()
	// [][3]int{转换后拼音字符串起始位置,转换后拼音长度,转换前字符数}
	hanzi := make([][3]int, 0)
	var text string
	var idx int
	for _, r := range raw {
		py := gpy.SinglePinyin(r, args)
		if len(py) > 0 {
			text += py[0]
			hanzi = append(hanzi, [3]int{
				idx,
				len(py[0]),
				utf8.RuneLen(r),
			})
			idx += len(py[0])
		} else {
			text += string(r)
			idx += len(string(r))
		}
	}
	return text, hanzi
}

// PinyinContains return matched substr index and length
func PinyinContains(raw string, term string) (int, int) {
	py, words := ToPinyin(raw)
	term, _ = ToPinyin(term)
	if !strings.Contains(py, term) {
		return -1, 0
	}
	py_bytes, term_bytes := []byte(py), []byte(term)
	for {
		idx := IndexBytes(py_bytes, term_bytes)
		if idx < 0 {
			break
		}
		valid := true
		for _, w := range words {
			if idx > w[0] && idx < w[0]+w[1] {
				py_bytes[idx] = 0
				valid = false
				break
			}
		}
		if valid {
			end := idx + len(term)
			for _, w := range words {
				if end-1 >= w[0] && end <= w[0]+w[1] {
					end = w[0] + w[1]
					break
				}
			}
			start_offset, end_offset := 0, 0
			for _, w := range words {
				if w[0] > end {
					break
				}
				if idx > w[0] {
					start_offset = start_offset + w[1] - w[2]
					end_offset = end_offset + w[1] - w[2]
				} else if w[0]+w[1] <= end {
					end_offset = end_offset + w[1] - w[2]
				}
			}
			return idx - start_offset, end - end_offset - (idx - start_offset)
		}
	}
	return -1, 0
}

// IndexBytes index
func IndexBytes(s []byte, term []byte) int {
	// TODO: use something like kmp someday
	for i := range s {
		found := true
		for j, tr := range term {
			if s[i+j] != tr {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}
