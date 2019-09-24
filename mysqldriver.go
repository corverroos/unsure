package unsure

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/go-sql-driver/mysql"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/log"
)

type conn struct {
	conn               driver.Conn
	connBeginTx        driver.ConnBeginTx
	connPrepareContext driver.ConnPrepareContext
	execerContext      driver.ExecerContext
	execer             driver.Execer
	queryerContext     driver.QueryerContext
	queryer            driver.Queryer
	pinger             driver.Pinger
}

func (c conn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare without context not supported")
}

func (c conn) Close() error {
	return c.conn.Close()
}

func (c conn) Begin() (driver.Tx, error) {
	return c.conn.Begin()
}

// Implement driver.ConnBeginTx
func (c conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	// Not tempting fate here since we don't always control it.
	return c.connBeginTx.BeginTx(ctx, opts)
}

// Implement driver.ConnPrepareContext
func (c conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if err := temptCtx(ctx); err != nil {
		return nil, err
	}
	return c.connPrepareContext.PrepareContext(ctx, query)
}

// Implement driver.Execer
func (c conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, errors.New("exec without context not supported")
}

// Implement driver.ExecerContext
func (c conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := temptCtx(ctx); err != nil {
		return nil, err
	}
	return c.execerContext.ExecContext(ctx, query, args)
}

// Implement driver.Queryer
func (c conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return nil, errors.New("query without context not supported")
}

// Implement driver.QueryerContext
func (c conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := temptCtx(ctx); err != nil {
		return nil, err
	}
	return c.queryerContext.QueryContext(ctx, query, args)
}

// Implement driver.Pinger
func (c conn) Ping(ctx context.Context) error {
	if err := temptCtx(ctx); err != nil {
		return err
	}
	return c.pinger.Ping(ctx)
}

// unsureDriver returns connections that always temp global fate.
type unsureDriver struct {
	md mysql.MySQLDriver
}

func (d unsureDriver) Open(dsn string) (driver.Conn, error) {
	c, err := d.md.Open(dsn)
	if err != nil {
		return nil, err
	}

	return conn{
		conn:               c,
		connBeginTx:        c.(driver.ConnBeginTx),
		connPrepareContext: c.(driver.ConnPrepareContext),
		execer:             c.(driver.Execer),
		execerContext:      c.(driver.ExecerContext),
		queryer:            c.(driver.Queryer),
		queryerContext:     c.(driver.QueryerContext),
		pinger:             c.(driver.Pinger),
	}, nil
}

func temptCtx(ctx context.Context) error {
	f, err := FateFromContext(ctx)
	if err != nil {
		log.Error(ctx, err)
		return err
	}
	return f.Tempt()
}

func init() {
	sql.Register(driverName, unsureDriver{})
}
