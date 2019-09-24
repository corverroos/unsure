package cursors

import (
	"database/sql"

	"github.com/luno/reflex"
	"github.com/luno/reflex/rsql"
)

// cursors wrap the engine_cursors table providing a reflex cursor store for any
// and all consumers running in engine.
var cursors = rsql.NewCursorsTable("engine_cursors",
	rsql.WithCursorCursorField("`cursor`"))

// ToStore returns a reflex cursor store backed by the engine_cursors table.
func ToStore(dbc *sql.DB) reflex.CursorStore {
	return cursors.ToStore(dbc, rsql.WithCursorAsyncDisabled()) // Have to disable async since it doesn't use fated context.
}
