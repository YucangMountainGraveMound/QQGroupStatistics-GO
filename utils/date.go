package utils

import (
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// FormatDateTime data string -> time.Time
func FormatDateTime(date, time string) time.Time {
	date = string([]rune(date)[4:])
	if strings.Contains(time, "上午") {
		time = string([]rune(time)[2:]) + " am"
	}
	if strings.Contains(time, "下午") {
		time = string([]rune(time)[2:]) + " pm"
	}
	d, err := dateparse.ParseAny(date + " " + time)
	if err != nil {
		panic(err)
	}
	return d
}
