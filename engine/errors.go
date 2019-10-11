package engine

import (
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

var (
	ErrActiveMatch       = errors.New("cannot start match, team has active match", j.C("ERR_3f2c52a21cf72a25"))
	ErrNoActiveMatch     = errors.New("team has no active match", j.C("ERR_87e5c79802c8ede2"))
	ErrOutOfSyncJoin     = errors.New("join in invalid state", j.C("ERR_799983ab899c3cdd"))
	ErrOutOfSyncSubmit   = errors.New("submit in invalid state", j.C("ERR_0d801e7015f73b13"))
	ErrOutOfSyncCollect  = errors.New("collect in invalid state", j.C("ERR_a8223c9740db7884"))
	ErrAlreadyJoined     = errors.New("player already joined round", j.C("ERR_9a15fbba1632ca47"))
	ErrAlreadyExcluded   = errors.New("player already excluded from round", j.C("ERR_9a15fbba1632ca47"))
	ErrUnknownPlayer     = errors.New("unknown player in round", j.C("ERR_7af903a349ee1789"))
	ErrIncorrectSubmit   = errors.New("submission answer incorrect", j.C("ERR_e823b34b4dd2a625"))
	ErrExcludedCollect   = errors.New("collect by excluded player", j.C("ERR_31f98641636f28dc"))
	ErrExcludedSubmit    = errors.New("submission by excluded player", j.C("ERR_8c5be353e22536fd"))
	ErrAlreadySubmitted  = errors.New("already submitted", j.C("ERR_95e55cdc7429f3a5"))
	ErrOutOfOrderSubmit  = errors.New("out of order submit", j.C("ERR_db1c92a67a7f6d8c"))
	ErrConcurrentUpdates = errors.New("concurrent round updates", j.C("ERR_b1294c9cc2be8b60"))
	ErrRoundNotFound     = errors.New("round not found for team", j.C("ERR_6139c6925dbcd93b"))
	ErrInactiveRound     = errors.New("round not part of active match", j.C("ERR_3a71fbdb00c931fd"))
)
