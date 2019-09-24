package ops

import "github.com/corverroos/unsure/engine"

//go:generate genbackendsimpl
type Backends interface {
	EngineClient() engine.Client
}
