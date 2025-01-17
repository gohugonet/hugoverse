package timestamp

import (
	"fmt"
	"strconv"
	"time"
)

func Now() string {
	return TimeToString(CurrentTimeMillis())
}

func TimeToString(time int64) string {
	return fmt.Sprintf("%d", time)
}

func CurrentTimeMillis() int64 {
	return int64(time.Nanosecond) * time.Now().UTC().UnixNano() / int64(time.Millisecond)
}

func TimeMillis(t time.Time) int64 {
	return int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)
}

func ConvertToTime(timestamp string) (time.Time, error) {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	tm := time.Unix(int64(i/1000), int64(i%1000))

	return tm, nil
}

func ConvertInt64ToTime(timestamp int64) (time.Time, error) {
	tm := time.Unix(timestamp/1000, timestamp%1000)

	return tm, nil
}
