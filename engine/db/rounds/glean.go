package rounds

import (
	"database/sql"

	"github.com/corverroos/unsure/engine/internal"
)

//go:generate glean -table=engine_rounds

type glean struct {
	internal.Round
	Error sql.NullString
}
