package rounds

import (
	"context"
	"database/sql"
	"time"

	"github.com/corverroos/unsure/engine"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/luno/shift"
)

func LookupByTeamAndIndex(ctx context.Context, dbc *sql.DB, team string,
	index int) (*internal.Round, error) {
	return lookupWhere(ctx, dbc, "team=? and `index`=?", team, index)
}

func ListByMatchID(ctx context.Context, dbc *sql.DB, matchID int64) ([]internal.Round, error) {
	return listWhere(ctx, dbc, "match_id=?", matchID)
}

func Join(ctx context.Context, dbc *sql.DB, team string, matchID int64, index int) error {
	_, err := fsm.Insert(ctx, dbc, joinReq{Team: team, MatchID: matchID, Index: index})
	return err
}

func ToJoined(ctx context.Context, dbc *sql.DB, id int64, from internal.RoundStatus,
	prevUpdatedAt time.Time, newState internal.RoundState) error {

	return to(ctx, dbc, id, from, internal.RoundStatusJoined, prevUpdatedAt,
		joinedReq{ID: id, State: newState})
}

func JoinedToCollect(ctx context.Context, dbc *sql.DB, id int64, prevUpdatedAt time.Time,
	newState internal.RoundState) error {

	return to(ctx, dbc, id, internal.RoundStatusJoined, internal.RoundStatusCollect,
		prevUpdatedAt, collectReq{ID: id, State: newState})
}

func ToCollected(ctx context.Context, dbc *sql.DB, id int64, from internal.RoundStatus,
	prevUpdatedAt time.Time, newState internal.RoundState) error {

	return to(ctx, dbc, id, from, internal.RoundStatusCollected, prevUpdatedAt,
		collectedReq{ID: id, State: newState})
}

func CollectedToSubmit(ctx context.Context, dbc *sql.DB, id int64, prevUpdatedAt time.Time) error {
	return to(ctx, dbc, id, internal.RoundStatusCollected, internal.RoundStatusSubmit,
		prevUpdatedAt, submitReq{ID: id})
}

func ToSubmitted(ctx context.Context, dbc *sql.DB, id int64, from internal.RoundStatus,
	prevUpdatedAt time.Time, newState internal.RoundState) error {

	return to(ctx, dbc, id, from, internal.RoundStatusSubmitted, prevUpdatedAt,
		submittedReq{ID: id, State: newState})
}

func SubmittedToSuccess(ctx context.Context, dbc *sql.DB, id int64, prevUpdatedAt time.Time) error {
	return to(ctx, dbc, id, internal.RoundStatusSubmitted, internal.RoundStatusSuccess,
		prevUpdatedAt, successReq{ID: id})
}

func ToFailed(ctx context.Context, dbc *sql.DB, id int64, from internal.RoundStatus,
	prevUpdatedAt time.Time, errMsg string) error {

	return to(ctx, dbc, id, from, internal.RoundStatusFailed, prevUpdatedAt,
		failedReq{ID: id, Error: errMsg})
}

func ensurePrevUpdatedAt(ctx context.Context, tx *sql.Tx, id int64, updatedAt time.Time) error {
	var n int
	err := tx.QueryRowContext(ctx, "select exists (select 1 from engine_rounds "+
		"where id=? and updated_at=?)", id, updatedAt).Scan(&n)
	if err != nil {
		return err
	}

	if n != 1 {
		return engine.ErrConcurrentUpdates
	}

	return nil
}

func to(ctx context.Context, dbc *sql.DB, id int64, from, to internal.RoundStatus,
	prevUpdatedAt time.Time, req shift.Updater) error {

	tx, err := dbc.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = ensurePrevUpdatedAt(ctx, tx, id, prevUpdatedAt)
	if err != nil {
		return err
	}

	notify, err := fsm.UpdateTx(ctx, tx, from, to, req)
	if err != nil {
		return err
	}
	defer notify()

	return tx.Commit()
}
