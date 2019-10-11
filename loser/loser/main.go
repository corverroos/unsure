package main

import (
	"github.com/corverroos/unsure"
	loser_ops "github.com/corverroos/unsure/loser/ops"
	"github.com/corverroos/unsure/loser/state"
	"github.com/luno/jettison/errors"
)

func main() {
	unsure.Bootstrap()

	s, err := state.New()
	if err != nil {
		unsure.Fatal(errors.Wrap(err, "new state error"))
	}

	loser_ops.StartLoops(s)

	unsure.WaitForShutdown()
}
