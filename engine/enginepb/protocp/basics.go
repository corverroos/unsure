package protocp

import (
	"math"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

var ErrOverflow = errors.New("integer conversion overflow", j.C("ERR_6d2270571432f6dd"))

func Uint32ToInt32(x uint32) (int32, error) {
	if x > math.MaxInt32 {
		return 0, ErrOverflow
	}

	return int32(x), nil
}

func Int32ToUint32(x int32) (uint32, error) {
	if x < 0 {
		return 0, ErrOverflow
	}

	return uint32(x), nil
}

func Uint64ToInt64(x uint64) (int64, error) {
	if x > math.MaxInt64 {
		return 0, ErrOverflow
	}

	return int64(x), nil
}

func Int64ToUint64(x int64) (uint64, error) {
	if x < 0 {
		return 0, ErrOverflow
	}

	return uint64(x), nil
}

func Uint32ToInt64(x uint32) (int64, error) {
	return int64(x), nil
}

func Int64ToUint32(x int64) (uint32, error) {
	if x < 0 || x > math.MaxUint32 {
		return 0, ErrOverflow
	}

	return uint32(x), nil
}

func Uint64ToInt32(x uint64) (int32, error) {
	if x > math.MaxInt32 {
		return 0, ErrOverflow
	}

	return int32(x), nil
}

func Int32ToUint64(x int32) (uint64, error) {
	if x < 0 {
		return 0, ErrOverflow
	}

	return uint64(x), nil
}
