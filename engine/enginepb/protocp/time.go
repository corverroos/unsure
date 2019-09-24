package protocp

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
)

func TimeToProto(t time.Time) (*timestamp.Timestamp, error) {
	return ptypes.TimestampProto(t)
}

func TimeFromProto(t *timestamp.Timestamp) (time.Time, error) {
	return ptypes.Timestamp(t)
}

func DurationToProto(d time.Duration) (*duration.Duration, error) {
	return ptypes.DurationProto(d), nil
}

func DurationFromProto(d *duration.Duration) (time.Duration, error) {
	return ptypes.Duration(d)
}

func TimeToProtoMs(t time.Time) (int64, error) {
	return t.UnixNano() / 1e6, nil
}

func TimeFromProtoMs(ms int64) (time.Time, error) {
	return time.Unix(0, ms*1e6), nil
}
