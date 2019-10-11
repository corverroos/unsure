package ops

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/corverroos/unsure/engine/db/matches"
	"github.com/corverroos/unsure/engine/db/rounds"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

// startRound idempotently starts a new round with index in the ready state.
func startRound(ctx context.Context, b Backends, matchID int64, index int) error {
	dbc := b.EngineDB().DB

	m, err := matches.Lookup(ctx, dbc, matchID)
	if err != nil {
		return err
	}

	_, err = rounds.LookupByTeamAndIndex(ctx, dbc, m.Team, index)
	if err == nil {
		// Round already started/exists.
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	// else err == sql.ErrNoRows

	return rounds.Join(ctx, dbc, m.Team, m.ID, index)
}

// maybeTimeoutRound fails a round (with error timeout) if it still is in status.
func maybeTimeoutRound(ctx context.Context, b Backends, roundID int64,
	status internal.RoundStatus) error {
	dbc := b.EngineDB().DB

	r, err := rounds.Lookup(ctx, dbc, roundID)
	if err != nil {
		return err
	}

	if r.Status != status {
		return nil
	}

	msg := fmt.Sprintf("status %s timeout", status.String())

	return rounds.ToFailed(ctx, dbc, r.ID, r.Status, r.UpdatedAt, msg)
}

// maybeCompleteMatch idempotently completes a match if enough rounds are complete.
func maybeCompleteMatch(ctx context.Context, b Backends, matchID int64) error {
	dbc := b.EngineDB().DB

	m, err := matches.Lookup(ctx, dbc, matchID)
	if err != nil {
		return err
	}

	if m.Status != internal.MatchStatusStarted {
		return nil
	}

	rl, err := rounds.ListByMatchID(ctx, dbc, m.ID)
	if err != nil {
		return err
	}

	if len(rl) < *roundCount {
		// Not all rounds started.
		return nil
	}

	first := time.Now()
	var last time.Time
	var success, failed int
	for _, r := range rl {
		if r.Status == internal.RoundStatusSuccess {
			success++
		} else if r.Status == internal.RoundStatusFailed {
			failed++
		} else {
			// Not all rounds complete, so just return.
			return nil
		}
		if r.UpdatedAt.After(last) {
			last = r.UpdatedAt
		}
		if r.CreatedAt.Before(first) {
			first = r.CreatedAt
		}
	}

	s := internal.MatchSummary{
		RoundsSuccess: success,
		RoundsFailed:  failed,
		Duration:      last.Sub(first),
	}

	return matches.EndMatch(ctx, dbc, matchID, s)
}

var advanceChecks = map[internal.RoundStatus]func(*internal.Match, *internal.Round) (bool, error){
	internal.RoundStatusJoined:    checkToCollect,
	internal.RoundStatusCollected: checkToSubmit,
	internal.RoundStatusSubmitted: checkToSuccess,
}
var advanceTos = map[internal.RoundStatus]func(ctx context.Context, dbc *sql.DB, r *internal.Round) error{
	internal.RoundStatusJoined:    advanceToCollect,
	internal.RoundStatusCollected: advanceToSubmit,
	internal.RoundStatusSubmitted: advanceToSuccess,
}

// maybeAdvanceRound fails a round (with error timeout) if it still is in status.
func maybeAdvanceRound(ctx context.Context, b Backends, roundID int64,
	status internal.RoundStatus) error {
	dbc := b.EngineDB().DB

	r, err := rounds.Lookup(ctx, dbc, roundID)
	if err != nil {
		return err
	}

	if r.Status != status {
		return nil
	}

	m, err := matches.Lookup(ctx, dbc, r.MatchID)
	if err != nil {
		return err
	}

	check, ok := advanceChecks[status]
	if !ok {
		return errors.New("no check defined for status", j.KV("status", status))
	}

	ok, err = check(m, r)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	advanceTo, ok := advanceTos[status]
	if !ok {
		return errors.New("no advance defined for status", j.KV("status", status))
	}

	return advanceTo(ctx, dbc, r)
}

func advanceToCollect(ctx context.Context, dbc *sql.DB, r *internal.Round) error {
	s := r.State

	genParts := func() map[string]int {
		parts := make(map[string]int)
		for _, m := range s.Players {
			parts[m.Name] = rand.Intn(100)
		}
		return parts
	}

	for i := 0; i < len(s.Players); i++ {
		s.Players[i].Parts = genParts()
	}

	return rounds.JoinedToCollect(ctx, dbc, r.ID, r.UpdatedAt, s)
}

func advanceToSubmit(ctx context.Context, dbc *sql.DB, r *internal.Round) error {
	return rounds.CollectedToSubmit(ctx, dbc, r.ID, r.UpdatedAt)
}

func advanceToSuccess(ctx context.Context, dbc *sql.DB, r *internal.Round) error {
	return rounds.SubmittedToSuccess(ctx, dbc, r.ID, r.UpdatedAt)
}

func checkToCollect(m *internal.Match, r *internal.Round) (bool, error) {
	return m.Players == len(r.State.Players), nil
}

func checkToSubmit(_ *internal.Match, r *internal.Round) (bool, error) {
	s := r.State

	for _, m := range s.Players {
		if !m.Collected {
			return false, nil
		}
	}

	return len(s.Players) > 0, nil
}

func checkToSuccess(_ *internal.Match, r *internal.Round) (bool, error) {
	s := r.State

	for _, m := range s.Players {
		if !m.Submitted {
			return false, nil
		}
	}

	return len(s.Players) > 0, nil
}
