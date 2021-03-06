package engine

import (
	"context"

	"github.com/luno/reflex"
)

// Client defines the root engine service interface.
type Client interface {
	Ping(context.Context) error
	Stream(ctx context.Context, after string, opts ...reflex.StreamOption) (reflex.StreamClient, error)
	StartMatch(ctx context.Context, team string, players int) error
	JoinRound(ctx context.Context, team string, player string, roundID int64) (bool, error)
	CollectRound(ctx context.Context, team string, player string, roundID int64) (*CollectRoundRes, error)
	SubmitRound(ctx context.Context, team string, player string, roundID int64, total int) error
}
