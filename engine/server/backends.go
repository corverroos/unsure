package server

import "github.com/corverroos/unsure/engine/db"

//go:generate genbackendsimpl
type Backends interface {
	EngineDB() *db.EngineDB
}
