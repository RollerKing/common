package crumbs

import (
	"time"
)

const (
	TIME_LAYOUT      = "2006-01-02 15:04:05"
	TIME_LAYOUT_LONG = "2006-01-02 15:04:05.999999"
	TIME_LAYOUT_DATE = "2006-01-02"
	DATE_LAYOUT      = "20060102"
	TIME_LAYOUT_TIME = "15:04:05"
)

// OptParse parse time without error
func OptParse(layout, value string) time.Time {
	tm, _ := time.ParseInLocation(layout, value, time.Now().Location())
	return tm
}

// OptParseFullTime parse 2006-01-02 15:04:05
func OptParseFullTime(value string) time.Time {
	return OptParse(TIME_LAYOUT, value)
}

// OptParseDate parse 2005-01-23
func OptParseDate(value string) time.Time {
	return OptParse(TIME_LAYOUT_DATE, value)
}

// OptParseTime parse 15:04:05
func OptParseTime(value string) time.Time {
	return OptParse(TIME_LAYOUT_TIME, value)
}
