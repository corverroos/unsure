package db

import (
	"database/sql"
	"flag"
	"runtime"
	"strings"
	"testing"

	"github.com/corverroos/unsure"
	"github.com/luno/jettison/log"
)

var (
	dbURI = flag.String("engine_db", "mysql://root@unix("+unsure.SockFile()+")/engine?",
		"engine DB URI")
)

type EngineDB struct {
	DB        *sql.DB
	ReplicaDB *sql.DB
}

// ReplicaOrMaster returns the replica DB if available, otherwise the master.
func (db *EngineDB) ReplicaOrMaster() *sql.DB {
	if db.ReplicaDB != nil {
		return db.ReplicaDB
	}
	return db.DB
}

func Connect() (*EngineDB, error) {
	ok, err := unsure.MaybeRecreateSchema(*dbURI, getSchemaPath())
	if err != nil {
		return nil, err
	} else if ok {
		log.Info(nil, "recreated schema")
	}

	dbc, err := unsure.Connect(*dbURI)
	if err != nil {
		return nil, err
	}
	return &EngineDB{
		DB:        dbc,
		ReplicaDB: dbc,
	}, nil
}

func ConnectForTesting(t *testing.T) *sql.DB {
	return unsure.ConnectForTesting(t, getSchemaPath())
}

func getSchemaPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return strings.Replace(filename, "connect.go", "schema.sql", 1)
}
