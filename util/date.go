package util

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func EndOfWeek(d time.Time) time.Time {
	endOfDay := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
	endOfDay = endOfDay.AddDate(0, 0, 1)
	endOfDay = endOfDay.Add(-1)
	for {
		if endOfDay.Weekday() == time.Sunday {
			return endOfDay
		} else {
			endOfDay = endOfDay.AddDate(0, 0, 1)
		}
	}
}

func SecondToDuration(secs int) string {
	var hours = int(math.Floor(float64(secs) / float64(60*60)))
	var minutes = int(math.Floor(float64(secs-(hours*60*60)) / float64(60)))
	var seconds = secs - (hours * 60 * 60) - (minutes * 60)
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func SameWeek(d1, d2 time.Time) bool {
	y1, w1 := d1.ISOWeek()
	y2, w2 := d2.ISOWeek()
	return y1 == y2 && w1 == w2
}

func DurationToSecond(dur string) int {
	var tmps = strings.Split(dur, ":")

	if len(tmps) == 3 {
		hours, _ := strconv.ParseInt(tmps[0], 10, 64)
		minutes, _ := strconv.ParseInt(tmps[1], 10, 64)
		seconds, _ := strconv.ParseInt(tmps[2], 10, 64)
		return int(hours*60*60) + int(minutes*60) + int(seconds)
	}

	if len(tmps) == 2 {
		minutes, _ := strconv.ParseInt(tmps[0], 10, 64)
		seconds, _ := strconv.ParseInt(tmps[1], 10, 64)
		return int(minutes*60) + int(seconds)
	}

	if len(tmps) == 1 {
		seconds, _ := strconv.ParseInt(tmps[0], 10, 64)
		return int(seconds)
	}

	return 0
}

func StartOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func EndOfDay(d time.Time) time.Time {
	var end = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
	return end.AddDate(0, 0, 1).Add(-1)
}

func ToUnixTime(d *time.Time) *int64 {
	if d != nil {
		var tmp = d.Unix()
		return &tmp
	} else {
		return nil
	}
}

func Now() time.Time {
	if os.Getenv("TIMEZONE") != "" {
		return time.Now().Add(time.Duration(-7) * time.Hour)
	}
	hours, _ := strconv.Atoi(os.Getenv("TIMEZONE"))
	return time.Now().Add(time.Duration(hours) * time.Hour)
}
