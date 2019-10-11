package ops

import (
	"database/sql"
	"fmt"

	"github.com/corverroos/unsure/engine"
	"github.com/corverroos/unsure/engine/db/rounds"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"golang.org/x/net/context"
)

// roundFailErrors defines the errors that result in round failure.
var roundFailErrors = map[error]bool{
	engine.ErrActiveMatch:       false,
	engine.ErrNoActiveMatch:     false,
	engine.ErrOutOfSyncJoin:     true,
	engine.ErrOutOfSyncSubmit:   true,
	engine.ErrOutOfSyncCollect:  true,
	engine.ErrAlreadyJoined:     false,
	engine.ErrUnknownPlayer:     true,
	engine.ErrIncorrectSubmit:   true,
	engine.ErrExcludedCollect:   true,
	engine.ErrExcludedSubmit:    true,
	engine.ErrAlreadySubmitted:  false,
	engine.ErrConcurrentUpdates: false,
	engine.ErrRoundNotFound:     false, // Cannot fail round not found
	engine.ErrInactiveRound:     false, // Cannot fail inactive round
}

type failer struct {
	ctx context.Context
	dbc *sql.DB
	r   *internal.Round
}

// newFailer returns a failer that can fail the round.
func newFailer(ctx context.Context, b Backends, r *internal.Round) *failer {
	return &failer{ctx, b.EngineDB().DB, r}
}

// err maybe fails the round and returns err.
func (f failer) err(err error) error {
	if err == nil {
		return nil
	}

	for ferr, fail := range roundFailErrors {
		if errors.Is(err, ferr) {
			if !fail {
				return err
			} else {
				break
			}
		}
	}
	msg := fmt.Sprintf("%+v", err)
	err2 := rounds.ToFailed(f.ctx, f.dbc, f.r.ID, f.r.Status, f.r.UpdatedAt, msg)
	if err2 != nil {
		return errors.Wrap(err2, "error failing round", j.KS("errMsg", err.Error()))
	}
	return err
}
