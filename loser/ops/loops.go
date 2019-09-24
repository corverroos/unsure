package ops

import (
	"context"
	"flag"
	"time"

	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/log"
)

var (
	team    = flag.String("team", "losers", "team name")
	player  = flag.String("player", "loser", "player name")
	players = flag.Int("players", 4, "number of players in the team")
)

func StartLoops(b Backends) {
	go startMatchForever(b)
}

func startMatchForever(b Backends) {
	for {
		ctx := unsure.ContextWithFate(context.Background(), unsure.DefaultFateP())

		err := b.EngineClient().StartMatch(ctx, *team, *players)

		if errors.Is(err, engine.ErrActiveMatch) {
			// Match active, just ignore
		} else if err != nil {
			log.Error(ctx, errors.Wrap(err, "start match error"))
		} else {
			log.Info(ctx, "match started")
		}

		time.Sleep(time.Second * 10)
	}
}
