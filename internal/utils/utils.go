package utils

import "time"

func GetTimeFromMillis(millis int64) time.Time {
	return time.UnixMilli(millis)
}
