package timex

import (
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
	_ "time/tzdata"
)

var (
	DefaultTimeZone = "Asia/Ho_Chi_Minh"
	LocalLocation   *time.Location
)

func init() {
	var err error
	LocalLocation, err = time.LoadLocation(DefaultTimeZone)
	if err != nil {
		logx.Errorf("failed to load time zone %q: %v", DefaultTimeZone, err)
	}
}

// Now returns the current time in the default location.
func Now() time.Time {
	return time.Now().In(LocalLocation)
}

// Parse parses a time string using the provided layout (in UTC).
func Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ParseInLocal parses a time string using the provided layout in the local location.
func ParseInLocal(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, LocalLocation)
}

// ParseStandard parses a datetime string using the standard layout in local time.
func ParseStandard(value string) (time.Time, error) {
	return ParseInLocal(time.DateTime, value)
}

// StartOfDay returns the start of the day (00:00:00) for a given time.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, LocalLocation)
}

// EndOfDay returns the end of the day (23:59:59) for a given time.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, LocalLocation)
}

// StartOfMonth returns the start time of the given month and year.
func StartOfMonth(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, LocalLocation)
}

// EndOfMonth returns the last moment of the given month and year.
func EndOfMonth(year, month int) time.Time {
	// Day 0 of next month = last day of this month
	return time.Date(year, time.Month(month)+1, 0, 23, 59, 59, 0, LocalLocation)
}

// IsBetween checks if checkTime is strictly between start and end.
func IsBetween(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

// IsAfter checks if the first time is after the second.
func IsAfter(a, b time.Time) bool {
	return a.After(b)
}

// IsBefore checks if the first time is before the second.
func IsBefore(a, b time.Time) bool {
	return a.Before(b)
}

// ConvertToTimeZone converts a time.Time to a different time zone.
func ConvertToTimeZone(t time.Time, zone string) (time.Time, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

// ProtoToTime converts a *timestamppb.Timestamp to *time.Time.
func ProtoToTime(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	tt := t.AsTime()
	return &tt
}

// TimeToProto converts a time.Time to *timestamppb.Timestamp.
func TimeToProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() || t.Unix() == 0 {
		return nil
	}
	return timestamppb.New(t)
}

// ProtoToTimeIn converts a protobuf Timestamp into a Go *time.Time,
// adjusted to the provided time zone. Returns nil if the input is nil.
func ProtoToTimeIn(ts *timestamppb.Timestamp, tz *time.Location) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime().In(tz)
	return &t
}

// ProtoToLocalTime converts a protobuf Timestamp into a Go *time.Time
// in the system's local time zone. Returns nil if the input is nil.
func ProtoToLocalTime(ts *timestamppb.Timestamp) *time.Time {
	return ProtoToTimeIn(ts, time.Local)
}
