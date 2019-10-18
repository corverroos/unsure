package unsure

import (
	"context"
	"time"

	"github.com/luno/fate"
	"github.com/luno/jettison"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/log"
	"github.com/luno/reflex"
)

// ConsumeForever is similar to rpatterns.ConsumeForever, but with shorter backoff.
func ConsumeForever(getCtx func() context.Context, consume reflex.ConsumeFunc,
	consumer reflex.Consumer, opts ...reflex.StreamOption) {
	for {
		ctx := getCtx()

		err := consume(ctx, consumer, opts...)
		if errors.IsAny(err, context.Canceled, context.DeadlineExceeded, reflex.ErrStopped, fate.ErrTempt) {
			reflexSoftErrorCounter.WithLabelValues(consumer.Name().String()).Inc()
			// Just retry on expected errors.
			time.Sleep(time.Millisecond * 100) // Don't spin
			continue
		}

		reflexHardErrorCounter.WithLabelValues(consumer.Name().String()).Inc()

		log.Error(ctx, errors.Wrap(err, "consume forever error"),
			jettison.WithKeyValueString("consumer", consumer.Name().String()))
		time.Sleep(time.Second) // 1 sec backoff on errors
	}
}
