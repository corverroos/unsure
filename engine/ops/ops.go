package ops

import (
	"context"
	"database/sql"
	"math/rand"

	"github.com/corverroos/unsure/engine"
	"github.com/corverroos/unsure/engine/db/matches"
	"github.com/corverroos/unsure/engine/db/rounds"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

func StartMatch(ctx context.Context, b Backends, team string, players int) error {
	_, err := matches.LookupActive(ctx, b.EngineDB().DB, team)
	if errors.Is(err, sql.ErrNoRows) {
		// no active match
	} else if err != nil {
		return err
	} else {
		return engine.ErrActiveMatch
	}

	_, err = matches.StartMatch(ctx, b.EngineDB().DB, team, players)
	return err
}

func JoinRound(ctx context.Context, b Backends, team, player string, roundID int64) (bool, error) {
	r, err := getRound(ctx, b, team, roundID)
	if err != nil {
		return false, err
	}

	m, err := matches.LookupActive(ctx, b.EngineDB().DB, team)
	if err != nil {
		return false, err
	}

	f := newFailer(ctx, b, r)

	if !internal.RoundStatusJoin.ThisOrNext(r.Status) {
		return false, f.err(engine.ErrNonReadyJoin)
	}

	s := r.State

	_, ms, ok := s.GetPlayer(player)
	if ok && ms.Included {
		return false, f.err(engine.ErrAlreadyJoined)
	} else if ok && !ms.Included {
		return false, f.err(engine.ErrAlreadyExcluded)
	}

	rank := rand.Intn(100)
	include := rand.Float64() > 0.5

	// If this is the lat player to join, ensure at least one is included.
	if len(s.Players) == m.Players-1 {
		var has bool
		for _, ps := range s.Players {
			if ps.Included {
				has = true
			}
		}
		if !has {
			include = true
		}
	}

	s.Players = append(s.Players, internal.RoundPlayerState{
		Name:     player,
		Included: include,
		Rank:     rank,
	})

	if err := ensureUniqRanks(s.Players); err != nil {
		return false, err
	}

	err = rounds.ToJoined(ctx, b.EngineDB().DB, r.ID, r.Status, r.UpdatedAt, s)
	if err != nil {
		return false, err
	}

	return include, nil
}

func ensureUniqRanks(states []internal.RoundPlayerState) error {
	uniq := make(map[int]bool)
	for _, s := range states {
		if uniq[s.Rank] {
			return errors.New("server error, generated duplicate rank")
		}
		uniq[s.Rank] = true
	}
	return nil
}

func getRound(ctx context.Context, b Backends, team string, roundID int64) (*internal.Round, error) {
	m, err := matches.LookupActive(ctx, b.EngineDB().DB, team)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, engine.ErrNoActiveMatch
	} else if err != nil {
		return nil, err
	}

	r, err := rounds.Lookup(ctx, b.EngineDB().DB, roundID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, engine.ErrRoundNotFound
	} else if err != nil {
		return nil, err
	}

	if r.Team != team {
		return nil, engine.ErrRoundNotFound
	} else if m.ID != r.MatchID {
		return nil, engine.ErrInactiveRound
	}

	return r, nil
}

func CollectRound(ctx context.Context, b Backends, team, player string, roundID int64) (*engine.CollectRoundRes, error) {
	r, err := getRound(ctx, b, team, roundID)
	if err != nil {
		return nil, err
	}

	f := newFailer(ctx, b, r)

	if !internal.RoundStatusCollect.ThisOrNext(r.Status) {
		return nil, f.err(engine.ErrNonSetDraw)
	}

	s := r.State

	n, ms, ok := s.GetPlayer(player)
	if !ok {
		return nil, f.err(engine.ErrUnknownPlayer)
	}

	if !ms.Included {
		return nil, f.err(engine.ErrNonIncludedDraw)
	}

	res := engine.CollectRoundRes{
		Rank: ms.Rank,
	}
	for _, m := range s.Players {
		res.Players = append(res.Players, engine.CollectPlayer{
			Name: m.Name,
			Part: m.Parts[m.Name],
		})
	}

	if !ms.Collected {
		s.Players[n].Collected = true
		err := rounds.ToCollected(ctx, b.EngineDB().DB, r.ID, r.Status, r.UpdatedAt, s)
		if err != nil {
			return nil, err
		}
	}

	return &res, nil
}

func SumbitRound(ctx context.Context, b Backends, team, player string, roundID int64, total int) (err error) {
	r, err := getRound(ctx, b, team, roundID)
	if err != nil {
		return err
	}

	f := newFailer(ctx, b, r)

	if !internal.RoundStatusSubmit.ThisOrNext(r.Status) {
		return f.err(engine.ErrNonGoSubmit)
	}

	s := r.State

	ml := s.GetSubmitOrder()

	var lastSubmitted internal.RoundPlayerState
	nextSubmit := ml[0]

	for i, m := range ml {
		if !m.Submitted {
			break
		}

		if nextSubmit.Name != m.Name {
			return errors.New("server error, invalid submit state")
		}

		lastSubmitted = m

		if i >= len(ml)-1 {
			nextSubmit = internal.RoundPlayerState{}
		} else {
			nextSubmit = ml[i+1]
		}
	}

	n, ms, ok := s.GetPlayer(player)
	if !ok {
		return f.err(engine.ErrUnknownPlayer)
	}

	if !ms.Included {
		return f.err(engine.ErrNonIncludedSubmit)
	}

	if total != s.GetTotal(player) {
		return f.err(engine.ErrIncorrectSubmit)
	}

	if ms.Name == lastSubmitted.Name {
		return f.err(engine.ErrAlreadySubmitted)
	}

	if ms.Name != nextSubmit.Name {
		return f.err(errors.Wrap(engine.ErrOutOfOrderSubmit, "wrap",
			j.MKS{"got": player, "want": nextSubmit.Name}))
	}

	s.Players[n].Submitted = true

	return rounds.ToSubmitted(ctx, b.EngineDB().DB, r.ID, r.Status, r.UpdatedAt, s)
}
