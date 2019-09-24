package events

import (
	"context"
	"database/sql"

	"github.com/luno/reflex"
	"github.com/luno/reflex/rsql"
)

// Events reflex events table
var events = rsql.NewEventsTableInt("engine_events")

// Insert inserts a reflex event into the onfido events table
// and returns a notify function or error.
func Insert(ctx context.Context, tx *sql.Tx, foreignID int64,
	typ reflex.EventType) (func(), error) {
	return events.Insert(ctx, tx, foreignID, typ)
}

// ToStream returns a reflex stream for onfido events.
func ToStream(dbc *sql.DB) reflex.StreamFunc {
	return events.ToStream(dbc)
}

func GetTable() rsql.EventsTableInt {
	return events
}
