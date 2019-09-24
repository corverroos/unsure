package rounds

import (
	"github.com/corverroos/unsure/engine/db/events"
	"github.com/corverroos/unsure/engine/internal"
	"github.com/luno/shift"
)

//go:generate shiftgen -inserter=joinReq -updaters=joinedReq,collectReq,collectedReq,submitReq,submittedReq,successReq,failedReq -table=engine_rounds

var fsm = shift.NewFSM(events.GetTable()).
	Insert(internal.RoundStatusJoin, joinReq{},
		internal.RoundStatusJoined, internal.RoundStatusFailed).
	Update(internal.RoundStatusJoined, joinedReq{},
		internal.RoundStatusJoined, internal.RoundStatusCollect, internal.RoundStatusFailed).
	Update(internal.RoundStatusCollect, collectReq{},
		internal.RoundStatusCollected, internal.RoundStatusFailed).
	Update(internal.RoundStatusCollected, collectedReq{},
		internal.RoundStatusCollected, internal.RoundStatusSubmit, internal.RoundStatusFailed).
	Update(internal.RoundStatusSubmit, submitReq{},
		internal.RoundStatusSubmitted, internal.RoundStatusFailed).
	Update(internal.RoundStatusSubmitted, submittedReq{},
		internal.RoundStatusSubmitted, internal.RoundStatusSuccess, internal.RoundStatusFailed).
	Update(internal.RoundStatusSuccess, successReq{}).
	Update(internal.RoundStatusFailed, failedReq{}).
	Build()

type joinReq struct {
	MatchID int64
	Index   int
	Team    string
}

type joinedReq struct {
	ID    int64
	State internal.RoundState
}

type collectReq struct {
	ID    int64
	State internal.RoundState
}

type collectedReq struct {
	ID    int64
	State internal.RoundState
}

type submitReq struct {
	ID int64
}

type submittedReq struct {
	ID    int64
	State internal.RoundState
}

type successReq struct {
	ID int64
}

type failedReq struct {
	ID    int64
	Error string
}
