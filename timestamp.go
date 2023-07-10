package fuse

import (
	"fmt"
	"time"
)

// Timestamp represents a Discord timestamp in milliseconds
type Timestamp struct {
	time.Time
}

func CreateTimestamp(t time.Time) Timestamp {
	return Timestamp{t}
}

func (t Timestamp) RelativeString() string {
	return fmt.Sprintf("<t:%d:R>", t.Unix())
}

// String returns a string representation of the timestamp in the format <t:1234567890> or <t:1234567890:R> if relative is true
func (t Timestamp) String() string {
	return fmt.Sprintf("<t:%d>", t.Unix())
}
