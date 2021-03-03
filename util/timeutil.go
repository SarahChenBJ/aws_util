package util

import (
	"fmt"
	"time"
)

const (
	//TimeFormat is for YYYY-MM-DD HH:MM:SS
	TimeFormat = "2006-01-02 15:04:05"
	//DateFormat is for YYYY-MM-DD
	DateFormat = "2006-01-02"
	//TimeFormatUnderline is for YYYY_MM_DD_HH_MM_SS
	TimeFormatUnderline = "2006_01_02_15_04_05"
	//YYYYMM formats for year and month
	YYYYMM = "2006-01"
)

//GetTime is a common method via format
func GetTime(t time.Time, format string) string {
	return t.Format(format)
}

//GetTimeByLocation return a time as string based on location
func GetTimeByLocation(t time.Time, loc *time.Location, format string) string {
	if loc != nil {
		return GetTime(t.In(loc), format)
	}
	return GetTime(t, format)
}

//GetDateByFormat get date string by custom format, input default format is DateFormat (from db)
func GetDateByFormat(str, format string) string {
	t, err := time.Parse(DateFormat, str)
	if err != nil {
		fmt.Errorf("GetDateByFormat error: ", err)
		return ""
	}
	return t.Format(format)
}

//StringToTime convert string to Time yyyy-mm-dd hh:mm:ss
func StringToTime(str string) (time.Time, error) {
	t, err := time.Parse(TimeFormat, str)
	return t, err
}

//StringToDate convert string to date yyyy-mm-dd
func StringToDate(str string) (time.Time, error) {
	t, err := time.Parse(DateFormat, str)
	return t, err
}

//LastdayOfMonth get the lastday of month by Time
func LastdayOfMonth(t time.Time) time.Time {
	first := FirstdayOfMonth(t)
	return first.AddDate(0, 1, -1)
}

//FirstdayOfMonth get the firstday of month by Time
func FirstdayOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

//TimestampToTime convert timestamp to Time
func TimestampToTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

//TimeToString convert Time to string
func TimeToString(t time.Time) string {
	return t.Format(TimeFormat)
}

//DateToString convert Date to string
func DateToString(t time.Time) string {
	return t.Format(DateFormat)
}

//MaxTime return maximum time between t1 and t2
func MaxTime(t1 time.Time, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t2
	}
	return t1
}

//MinTime return minimum time between t1 and t2
func MinTime(t1 time.Time, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}
	return t2
}

//Now get current time string
func Now() string {
	return time.Now().Format(TimeFormat)
}

//NowWithUnderLine get current time string separate by underline
func NowWithUnderLine() string {
	return time.Now().Format(TimeFormatUnderline)

}

// TimeToDate convert time to date
func TimeToDate(t time.Time) (time.Time, error) {
	str := DateToString(t)
	return StringToDate(str)
}

/*CompareStrDate  returns 2 dates compare result
** return > 0: d1 is after d2
** return = 0: d1 equal d2
** return < 0: d1 is before d2 or cannot compare
 */
func CompareStrDate(d1, d2 string) int {
	date1, e1 := StringToDate(d1)
	date2, e2 := StringToDate(d2)
	if e1 != nil || e2 != nil {
		return -1
	}
	if date1.After(date2) {
		return 1
	}
	if date1.Equal(date2) {
		return 0
	}
	return -1
}
