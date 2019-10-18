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
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"github.com/luno/reflex"
	"golang.org/x/sync/errgroup"
)

const (
	roundStatusTimeout = time.Minute // Timeout a round after this duration in single state.

	consumerStartRound    = "engine_start_rounds"
	consumerTimeoutRound  = "engine_timeout_round_%s"
	consumerAdvanceRound  = "engine_advance_round_%s"
	consumerMatchComplete = "engine_complete_match"
	consumerExitOnMatch   = "engine_exit_match_end"
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
		makeExitOnEnded(),
		makeStartRounds(b, *roundCount),
	}

	// Internal engine events consumable.
	consumable := reflex.NewConsumable(
		events.ToStream(b.EngineDB().DB),
		cursors.ToStore(b.EngineDB().DB))

	for _, req := range reqs {
		startConsume(b, consumable, req)
	}
}

// makeExitOnEnded returns a consumeReq that exists on match ended.
func makeExitOnEnded() consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, engine.EventTypeMatchEnded) {
			return nil
		}
		unsure.Fatal(errors.New("exit on match ended"))
		return nil
	}

	return newConsumeReq(consumerExitOnMatch, f)
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

// makeStartRounds returns a consumeReq that starts a new round (n) after
// a random delay after a MatchStarted event.
func makeStartRounds(b Backends, count int) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, engine.EventTypeMatchStarted) {
			return nil
		}

		var eg errgroup.Group
		for i := 0; i < count; i++ {
			ii := i

			// Start rounds in random order
			eg.Go(func() error {
				millis := rand.Intn(1000)
				time.Sleep(time.Millisecond * time.Duration(millis))

				if err := startRound(ctx, b, e.ForeignIDInt(), ii); err != nil {
					return err
				}

				log.Info(ctx, "round started", j.MKV{"index": ii})

				return fate.Tempt()
			})
		}

		err := eg.Wait()
		if err != nil {
			return err
		}

		return fate.Tempt()
	}

	name := reflex.ConsumerName(consumerStartRound)

	return newConsumeReq(name, f)
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
