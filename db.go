package unsure

import (
	"context"
	"database/sql"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/xo/dburl"
)

const (
	driverName = "unsuremysql"
)

var (
	dbTestURI  = flag.String("db_test_base", "mysql://root@unix("+SockFile()+")/test?", "Test database uri")
	dbRecreate = flag.Bool("db_recreate", false, "Whether to create the DB schema")

	dbMaxOpenConns    = flag.Int("db_max_open_conns", 100, "Maximum number of open database connections")
	dbMaxIdleConns    = flag.Int("db_max_idle_conns", 50, "Maximum number of idle database connections")
	dbConnMaxLifetime = flag.Duration("db_conn_max_lifetime", time.Minute, "Maximum time a single database connection can be left open")
)

func defaultOptions() string {
	// parseTime: Allows using time.Time for datetime
	// utf8mb4_general_ci: Needed for non-BMP unicode chars (e.g. emoji)
	return "parseTime=true&collation=utf8mb4_general_ci"
}

func SockFile() string {
	sock := "/tmp/mysql.sock"
	if _, err := os.Stat(sock); os.IsNotExist(err) {
		// try common linux/Ubuntu socket file location
		return "/var/run/mysqld/mysqld.sock"
	}
	return sock
}

func Connect(connectStr string) (*sql.DB, error) {
	const prefix = "mysql://"
	if !strings.HasPrefix(connectStr, prefix) {
		return nil, errors.New("db: URI is missing mysql:// prefix")
	}
	connectStr = connectStr[len(prefix):]

	if connectStr[len(connectStr)-1] != '?' {
		connectStr += "&"
	}
	connectStr += defaultOptions()

	dbc, err := sql.Open(driverName, connectStr)
	if err != nil {
		return nil, err
	}

	dbc.SetMaxOpenConns(*dbMaxOpenConns)
	dbc.SetMaxIdleConns(*dbMaxIdleConns)
	dbc.SetConnMaxLifetime(*dbConnMaxLifetime)

	return dbc, nil
}

// If db_name is empty, the default database will be used with temporary tables.
// Otherwise, database 'db_name' will be reset and used.
func ConnectForTesting(t testing.TB, schemaPaths ...string) *sql.DB {
	if len(schemaPaths) == 0 {
		t.Error("no schemas provided")
	}

	uri := *dbTestURI

	dbc, err := Connect(uri)
	if err != nil {
		t.Fatalf("connect error: %v", err)
		return nil
	}

	ctx := ContextWithFate(context.Background(), 0)

	// Multiple connections are problematic for unit tests since they
	// introduce concurrency issues.
	dbc.SetMaxOpenConns(1)

	if _, err := dbc.ExecContext(ctx, "set time_zone='+00:00';"); err != nil {
		t.Errorf("Error setting time_zone: %v", err)
	}
	_, err = dbc.ExecContext(ctx, "set sql_mode=if(@@version<'5.7', 'STRICT_TRANS_TABLES,NO_ENGINE_SUBSTITUTION', @@sql_mode);")
	if err != nil {
		t.Errorf("Error setting strict mode: %v", err)
	}

	for _, schemaPath := range schemaPaths {
		schema, err := ioutil.ReadFile(schemaPath)
		if err != nil {
			t.Errorf("Error reading schema: %s", err.Error())
			return nil
		}
		for _, q := range strings.Split(string(schema), ";") {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}

			q = strings.Replace(
				q, "create table", "create temporary table", 1)

			// Temporary tables don't support fulltext indexes.
			q = strings.Replace(
				q, "fulltext", "index", -1)

			_, err = dbc.ExecContext(ctx, q)
			if err != nil {
				t.Fatalf("Error executing %s: %s", q, err.Error())
				return nil
			}
		}
	}

	return dbc
}

func MaybeRecreateSchema(uri string, schemaPath string) (bool, error) {
	if !*dbRecreate {
		return false, nil
	}

	u, err := dburl.Parse(uri)
	if err != nil {
		return false, err
	}

	dbName := path.Base(u.Path)

	if err := execMysql("drop database if exists "+dbName, "", u); err != nil {
		return false, err
	}

	if err := execMysql("create database "+dbName, "", u); err != nil {
		return false, err
	}

	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return false, errors.Wrap(err, "error reading schema")
	}

	err = execMysql(string(schema), dbName, u)
	if err != nil {
		return false, err
	}

	return true, nil
}

func execMysql(stdIn string, db string, u *dburl.URL) error {
	var args []string
	if u.User != nil && u.User.Username() != "" {
		args = append(args, "-u", u.User.Username())
		if p, ok := u.User.Password(); ok && p != "" {
			args = append(args, "-p", p)
		}
	}
	args = append(args, db)
	cmd := exec.Command("mysql", args...)
	cmd.Stdin = strings.NewReader(stdIn)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "mysql error", j.KV("out", out))
	}
	return nil
}
