package rounds

import (
	"context"
	"testing"

	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine/db"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/stretchr/testify/require"
)

func TestFSM(t *testing.T) {
	dbc := db.ConnectForTesting(t)
	defer dbc.Close()

	ctx := unsure.ContextWithFate(context.Background(), 0)

	id, err := fsm.Insert(ctx, dbc, joinReq{
		Team:    "team",
		MatchID: 1,
		Index:   4,
	})
	require.NoError(t, err)

	r, err := Lookup(ctx, dbc, id)
	require.NoError(t, err)

	require.Equal(t, "team", r.Team)
	require.Equal(t, int64(4), r.Index)
	require.Equal(t, internal.RoundStatusJoin, r.Status)

	state := internal.RoundState{Players: []internal.RoundPlayerState{
		{
			Name:     "player1",
			Rank:     1,
			Included: false,
			Parts:    map[string]int{"2": 1},
		},
	}}

	err = ToJoined(ctx, dbc, id, internal.RoundStatusJoin, r.UpdatedAt, state)
	require.NoError(t, err)

	r, err = Lookup(ctx, dbc, id)
	require.NoError(t, err)

	require.Equal(t, "team", r.Team)
	require.Equal(t, int64(4), r.Index)
	require.Equal(t, internal.RoundStatusJoined, r.Status)
	require.Equal(t, state, r.State)
}
