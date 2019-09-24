package matches

import (
	"context"
	"database/sql"

	"github.com/corverroos/unsure/engine/internal"
)

func LookupActive(ctx context.Context, dbc *sql.DB, team string) (*internal.Match, error) {
	return lookupWhere(ctx, dbc, "team=? and status=?", team, internal.MatchStatusStarted)
}

func StartMatch(ctx context.Context, dbc *sql.DB, team string, players int) (int64, error) {
	return fsm.Insert(ctx, dbc, startReq{Team: team, Players: players})
}

func EndMatch(ctx context.Context, dbc *sql.DB, matchID int64,
	summary internal.MatchSummary) error {

	return fsm.Update(ctx, dbc, internal.MatchStatusStarted,
		internal.MatchStatusEnded, endReq{ID: matchID, Summary: summary})
}
