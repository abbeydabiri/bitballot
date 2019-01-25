package utils

import (
	"log"
	"math"
	"time"

	"fmt"

	"sort"
	"strconv"
	"bitballot/config"
)

func GetMonthPeriod(now time.Time) (firstDay, lastDay time.Time) {

	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstDay = time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastDay = firstDay.AddDate(0, 1, -1)

	return
}

// Seconds-based time units
const (
	Minute     = 60
	Hour       = 60 * Minute
	Day        = 24 * Hour
	Week       = 7 * Day
	Month      = 30 * Day
	Year       = 12 * Month
	LongTime   = 37 * Year
	TimeFormat = "02/01/2006 15:04:05"
)

var magnitudes = []struct {
	d      int64
	format string
	divby  int64
}{
	{1, "now", 1},
	{2, "1 second %s", 1},
	{Minute, "%d seconds %s", 1},
	{2 * Minute, "1 minute %s", 1},
	{Hour, "%d minutes %s", Minute},
	{2 * Hour, "1 hour %s", 1},
	{Day, "%d hours %s", Hour},
	{2 * Day, "1 day %s", 1},
	{Week, "%d days %s", Day},
	{2 * Week, "1 week %s", 1},
	{Month, "%d weeks %s", Week},
	{2 * Month, "1 month %s", 1},
	{Year, "%d months %s", Month},
	{18 * Month, "1 year %s", 1},
	{2 * Year, "2 years %s", 1},
	{LongTime, "%d years %s", Year},
	{math.MaxInt64, "a long while %s", 1},
}

// GetDuration formats a time into a relative string.
//
// It takes two times and two labels.  In addition to the generic time
// delta string (e.g. 5 minutes), the labels are used applied so that
// the label corresponding to the smaller time is applied.
//
// GetDuration(timeInPast, timeInFuture, "earlier", "later") -> "3 weeks earlier"
// func GetDuration(a, b time.Time, cLabelpast, cLabelfuture string) string {

func GetTime() time.Time {
	return time.Now()
}

func GetLocalTime() time.Time {
	curTime := time.Now()
	loc, err := time.LoadLocation(config.Get().Timezone)
	if err != nil {
		log.Println(err)
		return curTime
	}
	return curTime.In(loc)
}

func DateTimeMerge(curDate, curTime, location string) time.Time {
	if curTime == "" {
		curTime = "15:04"
	}

	loc, err := time.LoadLocation(location)
	if err != nil {
		log.Println(err)
		loc = time.UTC
	}

	tTime, err := time.ParseInLocation("2006-01-02 15:04:05",
		fmt.Sprintf("%v %v:05", curDate, curTime), loc)
	if err != nil {
		log.Printf(err.Error())
	}
	return tTime
}

func DateTimeSplit(tDateTime time.Time) (curDate, curTime string) {
	curDate = tDateTime.Format("2006-01-02")
	curTime = tDateTime.Format("15:04")
	return
}

func GetTimeHuman(curTime time.Time) string {
	return curTime.Format("02/01/2006 3:04 PM")
}

func GetTimeFormat(curTime time.Time) string {
	return curTime.Format("02/01/2006 15:04:05")
}

func GetSystemTime() string {
	curTime := GetLocalTime()
	return curTime.Format("02/01/2006 15:04:05")
}

func GetSystemDate() string {
	return time.Now().Format("02/01/2006")
}

func GetUnixString(cTime string) string {
	cFormat := "02/01/2006 15:04:05"
	tTime, err := time.Parse(cFormat, cTime)
	if err != nil {
		return ""
	}

	return strconv.FormatInt(tTime.Unix(), 10)
}

func GetUnixTimestamp() string {
	return strconv.FormatInt(time.Now().Local().Unix(), 10)
}

func GetDifferenceInYears(cTimeCur string, cTimePast string) (years int) {

	years = int(0)
	cFormat := "02/01/2006"

	if cTimeCur == "" {
		cTimeCur = time.Now().Format(cFormat)
	}
	tTimeCur, errCur := time.Parse(cFormat, cTimeCur)
	if errCur != nil {
		return
	}

	if cTimePast == "" {
		cTimePast = time.Now().Format(cFormat)
	}
	tTimePast, errPast := time.Parse(cFormat, cTimePast)
	if errPast != nil {
		return
	}

	if errCur == nil && errPast == nil {

		if tTimePast.After(tTimeCur) {
			tTimeCur, tTimePast = tTimePast, tTimeCur
		}

		yearCur, _, _ := tTimeCur.Date()
		yearPast, _, _ := tTimePast.Date()

		years = int(yearCur - yearPast)
	}

	return
}

func GetDifferenceInMonths(cTimeCur string, cTimePast string) (months int) {

	months = int(0)
	cFormat := "02/01/2006"

	if cTimeCur == "" {
		cTimeCur = time.Now().Format(cFormat)
	}
	tTimeCur, errCur := time.Parse(cFormat, cTimeCur)
	if errCur != nil {
		return
	}

	if cTimePast == "" {
		cTimePast = time.Now().Format(cFormat)
	}
	tTimePast, errPast := time.Parse(cFormat, cTimePast)
	if errPast != nil {
		return
	}

	if errCur == nil && errPast == nil {

		if tTimePast.After(tTimeCur) {
			tTimeCur, tTimePast = tTimePast, tTimeCur
		}

		_, monthCur, _ := tTimeCur.Date()
		_, monthPast, _ := tTimePast.Date()

		months = int(monthCur - monthPast)
	}

	return
}

func GetDifferenceInSeconds(cTimeCur string, cTimePast string) (seconds int) {

	seconds = int(0)
	cFormat := "Mon, 02 Jan 2006 15:04:05"

	if cTimeCur == "" {
		cTimeCur = time.Now().Format(cFormat)
	}
	tTimeCur, errCur := time.Parse(cFormat, cTimeCur)
	if errCur != nil {
		return
	}

	if cTimePast == "" {
		cTimePast = time.Now().Format(cFormat)
	}
	tTimePast, errPast := time.Parse(cFormat, cTimePast)
	if errPast != nil {
		return
	}

	if errCur == nil && errPast == nil {

		secondsCur := tTimeCur.Unix()
		secondsPast := tTimePast.Unix()
		seconds = int(secondsCur - secondsPast)
	}

	return
}

func GetDuration(cPast, cPresent string) string {
	cLabel := "ago"
	cFormat := "02/01/2006 15:04:05"

	tPast, err := time.Parse(cFormat, cPast)
	if err != nil {
		return ""
	}

	tPresent, err := time.Parse(cFormat, cPresent)
	if err != nil {
		return ""
	}

	diff := tPresent.Unix() - tPast.Unix()

	after := tPast.After(tPresent)
	if after {
		cLabel = "from now"
		diff = tPast.Unix() - tPresent.Unix()
	}

	n := sort.Search(len(magnitudes), func(i int) bool {
		return magnitudes[i].d > diff
	})

	mag := magnitudes[n]
	args := []interface{}{}
	escaped := false
	for _, ch := range mag.format {
		if escaped {
			switch ch {
			case '%':
			case 's':
				args = append(args, cLabel)
			case 'd':
				args = append(args, diff/mag.divby)
			}
			escaped = true
		} else {
			escaped = ch == '%'
		}
	}
	return fmt.Sprintf(mag.format, args...)
}
