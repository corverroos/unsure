package matches

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

	id, err := fsm.Insert(ctx, dbc, startReq{
		Team:    "team",
		Players: 4,
	})
	require.NoError(t, err)

	m, err := Lookup(ctx, dbc, id)
	require.NoError(t, err)

	require.Equal(t, "team", m.Team)
	require.Equal(t, 4, m.Players)
	require.Equal(t, internal.MatchStatusStarted, m.Status)

	err = fsm.Update(ctx, dbc, internal.MatchStatusStarted, internal.MatchStatusEnded, endReq{ID: id})
	require.NoError(t, err)

	m, err = Lookup(ctx, dbc, id)
	require.NoError(t, err)

	require.Equal(t, "team", m.Team)
	require.Equal(t, 4, m.Players)
	require.Equal(t, internal.MatchStatusEnded, m.Status)
}
