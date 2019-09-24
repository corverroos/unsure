package state

import "github.com/corverroos/unsure/engine/db"

type State struct {
	engineDB *db.EngineDB
}

func (s *State) EngineDB() *db.EngineDB {
	return s.engineDB
}

// New returns a new engine state.
func New() (*State, error) {
	var (
		s   State
		err error
	)

	s.engineDB, err = db.Connect()
	if err != nil {
		return nil, err
	}

	return &s, nil
}
