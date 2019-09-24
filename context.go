package unsure

import (
	"flag"
	"testing"

	"github.com/luno/fate"
	"github.com/luno/jettison/errors"
	"golang.org/x/net/context"
)

var (
	defaultFateP = flag.Float64("fate_p", 0.33, "default fate probability")
	cheatFate    = flag.Bool("cheat_fate", false, "cheat fate by not requiring fate in context")
)

func DefaultFateP() float64 {
	return *defaultFateP
}

type fateKey struct{}

func ContextWithFate(ctx context.Context, fateP float64) context.Context {
	f := fate.New(fate.WithDefaultP(fateP), fate.WithoutOfficeHours())
	return context.WithValue(ctx, fateKey{}, f)
}

func FateFromContext(ctx context.Context) (fate.Fate, error) {
	v := ctx.Value(fateKey{})
	if v == nil {
		if *cheatFate {
			return fate.New(fate.WithDefaultP(0), fate.WithoutOfficeHours()), nil
		}
		return nil, errors.New("context missing fate")
	}
	f, ok := v.(fate.Fate)
	if !ok {
		return nil, errors.New("invalid context fate")
	}
	return f, nil
}

func CheatFateForTesting(_ *testing.T) func() {
	cache := cheatFate
	*cheatFate = true
	return func() {
		cheatFate = cache
	}
}
