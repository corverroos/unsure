package unsure

import (
	"os"
	"sync"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/log"
)

// ShutdownFn is a shutdown function that can be added to a registry.
type ShutdownFn func() error

// Registry is a registry of shutdown functions.
type Registry struct {
	fns []ShutdownFn

	mu           sync.Mutex
	shuttingDown bool
}

// Register registers a shutdown function. Shutdown functions will be called in
// the order in which they're registered.
func (r *Registry) Register(fn ShutdownFn) {
	r.fns = append(r.fns, fn)
}

// Shutdown calls all registered shutdown functions and returns the first error
// encountered.
//
// TODO(neil): Add timeouts
func (r *Registry) Shutdown() error {

	r.mu.Lock()
	shuttingDown := r.shuttingDown
	r.shuttingDown = true
	r.mu.Unlock()

	if shuttingDown {
		return nil
	}

	var anyErr error
	for _, fn := range r.fns {
		if err := fn(); err != nil {
			log.Error(nil, errors.Wrap(err, "error running shutdown fn"))
			anyErr = err
		}
	}
	return anyErr
}

var defaultRegistry = new(Registry)

// Register registers a shutdown function with the default registry.
func RegisterShutdown(fn ShutdownFn) {
	defaultRegistry.Register(fn)
}

// RegisterNoErr registers a no-error shutdown function with the default registry.
func RegisterNoErr(fn func()) {
	RegisterShutdown(func() error {
		fn()
		return nil
	})
}

// Shutdown calls all shutdown functions in the default registry and exits with
// an appropriate error code.
func shutdown() {
	if err := defaultRegistry.Shutdown(); err != nil {
		log.Error(nil, errors.Wrap(err, "app shutdown with error"))
		os.Exit(1)
	}

	log.Info(nil, "app shut down gracefully")
	os.Exit(0)
}
