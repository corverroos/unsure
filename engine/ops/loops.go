package ops

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine"
	"github.com/corverroos/unsure/engine/db/cursors"
	"github.com/corverroos/unsure/engine/db/events"
	"github.com/corverroos/unsure/engine/db/rounds"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/luno/fate"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"github.com/luno/reflex"
)

const (
	roundStatusTimeout = time.Minute // Timeout a round after this duration in single state.

	consumerStartRound    = "engine_start_round_%d"
	consumerTimeoutRound  = "engine_timeout_round_%s"
	consumerAdvanceRound  = "engine_advance_round_%s"
	consumerMatchComplete = "engine_complete_match"
)

var (
	roundCount = flag.Int("rounds", 10, "number of rounds per match")
)

func StartLoops(b Backends) {
	reqs := []consumeReq{
		makeTimeoutRound(b, internal.RoundStatusJoin),
		makeTimeoutRound(b, internal.RoundStatusJoined),
		makeTimeoutRound(b, internal.RoundStatusCollect),
		makeTimeoutRound(b, internal.RoundStatusCollected),
		makeTimeoutRound(b, internal.RoundStatusSubmit),
		makeTimeoutRound(b, internal.RoundStatusSubmitted),

		makeAdvanceRound(b, internal.RoundStatusJoined),
		makeAdvanceRound(b, internal.RoundStatusCollected),
		makeAdvanceRound(b, internal.RoundStatusSubmitted),

		makeCompleteMatch(b),
	}

	for i := 0; i < *roundCount; i++ {
		reqs = append(reqs, makeStartRound(b, i))
	}

	// Internal engine events consumable.
	consumable := reflex.NewConsumable(
		events.ToStream(b.EngineDB().DB),
		cursors.ToStore(b.EngineDB().DB))

	for _, req := range reqs {
		startConsume(b, consumable, req)
	}
}

// makeAdvanceRound returns a consumeReq that times out a round if it too long in
// a specific state.
func makeCompleteMatch(b Backends) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsAnyType(e.Type, internal.RoundStatusSuccess, internal.RoundStatusFailed) {
			return nil
		}

		r, err := rounds.Lookup(ctx, b.EngineDB().DB, e.ForeignIDInt())
		if err != nil {
			return err
		}

		if err := maybeCompleteMatch(ctx, b, r.MatchID); err != nil {
			return err
		}

		return fate.Tempt()
	}

	return newConsumeReq(consumerMatchComplete, f)
}

// makeAdvanceRound returns a consumeReq that advances a round to its
// subsequent state if its checks pass.
func makeAdvanceRound(b Backends, status internal.RoundStatus) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, status) {
			return nil
		}

		if err := maybeAdvanceRound(ctx, b, e.ForeignIDInt(), status); err != nil {
			return err
		}

		return fate.Tempt()
	}

	name := reflex.ConsumerName(fmt.Sprintf(consumerAdvanceRound, status.String()))

	return newConsumeReq(name, f)
}

// makeTimeoutRound returns a consumeReq that times out a round if it too long in
// a specific state.
func makeTimeoutRound(b Backends, status internal.RoundStatus) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, status) {
			return nil
		}

		if err := maybeTimeoutRound(ctx, b, e.ForeignIDInt(), status); err != nil {
			return err
		}

		return fate.Tempt()
	}

	name := reflex.ConsumerName(fmt.Sprintf(consumerTimeoutRound, status.String()))

	return newConsumeReq(name, f, reflex.WithStreamLag(roundStatusTimeout))
}

// makeStartRound returns a consumeReq that starts a new round (n) after
// a random delay after a MatchStarted event.
func makeStartRound(b Backends, n int) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, engine.EventTypeMatchStarted) {
			return nil
		}

		if err := startRound(ctx, b, e.ForeignIDInt(), n); err != nil {
			return err
		}

		log.Info(ctx, "round started", j.MKV{"index": n})

		return fate.Tempt()
	}

	delay := rand.Intn(1 + (n * 1)) // Max n * 1 secs, min 1 sec.
	delayS := time.Second * time.Duration(delay)

	name := reflex.ConsumerName(fmt.Sprintf(consumerStartRound, n))

	return newConsumeReq(name, f, reflex.WithStreamLag(delayS))
}

type consumeReq struct {
	name  reflex.ConsumerName
	f     func(ctx context.Context, f fate.Fate, e *reflex.Event) error
	copts []reflex.ConsumerOption
	sopts []reflex.StreamOption
}

func newConsumeReq(name reflex.ConsumerName, f func(ctx context.Context, f fate.Fate, e *reflex.Event) error,
	opts ...reflex.StreamOption) consumeReq {
	return consumeReq{
		name:  name,
		f:     f,
		sopts: opts,
	}
}

func startConsume(b Backends, c reflex.Consumable, req consumeReq) {
	consumer := reflex.NewConsumer(req.name, req.f, req.copts...)
	go unsure.ConsumeForever(unsure.FatedContext, c.Consume, consumer, req.sopts...)
}
