package protocp

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
)

func NullInt64ToProto(n sql.NullInt64) (int64, error) {
	return n.Int64, nil
}

func NullInt64FromProto(n int64) (sql.NullInt64, error) {
	return sql.NullInt64{
		Valid: true,
		Int64: n,
	}, nil
}

func NullStringToProto(s sql.NullString) (string, error) {
	return s.String, nil
}

func NullStringFromProto(s string) (sql.NullString, error) {
	return sql.NullString{
		Valid:  true,
		String: s,
	}, nil
}

func NullBoolToProto(b sql.NullBool) (bool, error) {
	return b.Bool, nil
}

func NullBoolFromProto(b bool) (sql.NullBool, error) {
	return sql.NullBool{
		Valid: true,
		Bool:  b,
	}, nil
}

func NullFloat64ToProto(f sql.NullFloat64) (float64, error) {
	return f.Float64, nil
}

func NullFloat64FromProto(f float64) (sql.NullFloat64, error) {
	return sql.NullFloat64{
		Valid:   true,
		Float64: f,
	}, nil
}

func NullTimeToProtoMs(t mysql.NullTime) (int64, error) {
	return t.Time.UnixNano() / 1e6, nil
}

func NullTimeFromProtoMs(ms int64) (mysql.NullTime, error) {
	return mysql.NullTime{
		Valid: true,
		Time:  time.Unix(0, ms*1e6),
	}, nil
}
