package unsure_test

import (
	"context"
	"testing"

	"github.com/corverroos/unsure"
	"github.com/stretchr/testify/require"
)

func TestCtxFate(t *testing.T) {
	ctx := unsure.ContextWithFate(context.Background(), 1)
	fate, err := unsure.FateFromContext(ctx)
	require.NoError(t, err)
	require.Error(t, fate.Tempt())

	_, err = unsure.FateFromContext(context.Background())
	require.Error(t, err)
}
