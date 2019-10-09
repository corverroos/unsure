package engine

import (
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

var (
	ErrActiveMatch       = errors.New("cannot start match, team has active match", j.C("ERR_3f2c52a21cf72a25"))
	ErrNoActiveMatch     = errors.New("team has no active match", j.C("ERR_87e5c79802c8ede2"))
	ErrNonReadyJoin      = errors.New("joining when round not ready", j.C("ERR_799983ab899c3cdd"))
	ErrNonGoSubmit       = errors.New("submitting when round not go", j.C("ERR_0d801e7015f73b13"))
	ErrNonSetDraw        = errors.New("drawing when round not set", j.C("ERR_a8223c9740db7884"))
	ErrAlreadyJoined     = errors.New("player already joined round", j.C("ERR_9a15fbba1632ca47"))
	ErrUnknownPlayer     = errors.New("unknown player in round", j.C("ERR_7af903a349ee1789"))
	ErrOutOfOrderSubmit  = errors.New("out of order submission", j.C("ERR_5401aa4803dde9c3"))
	ErrIncorrectSubmit   = errors.New("submission answer incorrect", j.C("ERR_e823b34b4dd2a625"))
	ErrNonIncludedDraw   = errors.New("draw by non-joined player", j.C("ERR_31f98641636f28dc"))
	ErrNonIncludedSubmit = errors.New("submission by non-joined player", j.C("ERR_8c5be353e22536fd"))
	ErrAlreadySubmitted  = errors.New("already submitted", j.C("ERR_95e55cdc7429f3a5"))
	ErrConcurrentUpdates = errors.New("concurrent round updates", j.C("ERR_b1294c9cc2be8b60"))
	ErrRoundNotFound     = errors.New("round not found for team", j.C("ERR_6139c6925dbcd93b"))
	ErrInactiveRound     = errors.New("round not part of active match", j.C("ERR_3a71fbdb00c931fd"))
)
