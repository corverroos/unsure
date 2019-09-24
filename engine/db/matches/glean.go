package matches

import (
	"github.com/corverroos/unsure/engine/internal"
)

//go:generate glean -table=engine_matches

type glean struct {
	internal.Match
}
