package state

import (
	"github.com/corverroos/unsure/engine"
	ec "github.com/corverroos/unsure/engine/client"
)

type State struct {
	engineClient engine.Client
}

func (s *State) EngineClient() engine.Client {
	return s.engineClient
}

// New returns a new engine state.
func New() (*State, error) {
	var (
		s   State
		err error
	)

	s.engineClient, err = ec.New()
	if err != nil {
		return nil, err
	}

	return &s, nil
}
