package ops

import (
	"context"
	"flag"
	"time"

	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"github.com/luno/reflex"
)

var (
	team    = flag.String("team", "losers", "team name")
	player  = flag.String("player", "loser", "player name")
	players = flag.Int("players", 4, "number of players in the team")
)

func StartLoops(b Backends) {
	go startMatchForever(b)
	go logHeadForever(b)
}

func logHeadForever(b Backends) {
	var delay time.Duration
	for {
		time.Sleep(delay)
		delay = time.Second

		ctx := unsure.FatedContext()
		cl, err := b.EngineClient().Stream(ctx, "", reflex.WithStreamFromHead())
		if err != nil {
			log.Error(ctx, errors.Wrap(err, "log head stream error"))
			continue
		}
		for {
			e, err := cl.Recv()
			if err != nil {
				log.Error(ctx, errors.Wrap(err, "log head recv error"))
				break
			}
			typ := engine.EventType(e.Type.ReflexType())
			log.Info(ctx, "log head consumed event",
				j.MKV{"id": e.ForeignIDInt(), "type": typ, "latency": time.Since(e.Timestamp)})
		}
	}
}

func startMatchForever(b Backends) {
	for {
		ctx := unsure.ContextWithFate(context.Background(), unsure.DefaultFateP())

		err := b.EngineClient().StartMatch(ctx, *team, *players)

		if errors.Is(err, engine.ErrActiveMatch) {
			// Match active, just ignore
			return
		} else if err != nil {
			log.Error(ctx, errors.Wrap(err, "start match error"))
		} else {
			log.Info(ctx, "match started")
			return
		}

		time.Sleep(time.Second)
	}
}
