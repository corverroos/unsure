package matches

import (
	"github.com/corverroos/unsure/engine/db/events"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/luno/shift"
)

//go:generate shiftgen -inserter=startReq -updaters=endReq -table=engine_matches

var fsm = shift.NewFSM(events.GetTable()).
	Insert(internal.MatchStatusStarted, startReq{}, internal.MatchStatusEnded).
	Update(internal.MatchStatusEnded, endReq{}).
	Build()

type startReq struct {
	Team    string
	Players int
}

type endReq struct {
	ID      int64
	Summary internal.MatchSummary
}
