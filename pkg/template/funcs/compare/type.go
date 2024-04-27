package compare

import "time"

type TimeZone interface {
	Location() *time.Location
}
