package utils

import "time"

func DateOnly(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func Yesterday(now time.Time) time.Time {
	return now.AddDate(0, 0, -1)
}
