package await_signals

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

/**
 * The sample demonstrates how to deal with multiple signals that can come out of order and require actions
 * if a certain signal not received in a specified time interval.
 *
 * This specific sample receives three signals: Signal1, Signal2, Signal3. They have to be processed in the
 * sequential order, but they can be received out of order.
 * There are two timeouts to enforce.
 * The first one is the maximum time between signals.
 * The second limits the total time since the first signal received.
 *
 *  The following example demonstrates an approach. It waits on each signal receive channel in turn.
 */

// SignalToSignalTimeout is the maximum time between signals
var SignalToSignalTimeout = 30 * time.Second

// FromFirstSignalTimeout is the maximum time to receive all signals
var FromFirstSignalTimeout = 60 * time.Second

type AwaitSignals struct {
	FirstSignalTime time.Time
	Signal1Received bool
	Signal2Received bool
	Signal3Received bool
}

// GetNextTimeout returns the maximum time allowed to wait for the next signal.
func (a *AwaitSignals) GetNextTimeout(ctx workflow.Context) (time.Duration, error) {
	if a.FirstSignalTime.IsZero() {
		panic("FirstSignalTime is not yet set")
	}
	total := workflow.Now(ctx).Sub(a.FirstSignalTime)
	totalLeft := FromFirstSignalTimeout - total
	if totalLeft <= 0 {
		return 0, temporal.NewApplicationError("FromFirstSignalTimeout", "timeout")
	}
	if SignalToSignalTimeout < totalLeft {
		return SignalToSignalTimeout, nil
	}
	return totalLeft, nil
}

// AwaitSignalsWorkflow workflow definition
func AwaitSignalsWorkflow(ctx workflow.Context) (err error) {
	log := workflow.GetLogger(ctx)
	var a AwaitSignals
	receivedFirstSignal := false
	timedOut := false
	for _, signal := range []string{"Signal1", "Signal2", "Signal3"} {
		s := workflow.NewSelector(ctx)
		s.AddReceive(workflow.GetSignalChannel(ctx, signal), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			log.Info(signal + " processed")
		})
		if receivedFirstSignal {
			timeout, err := a.GetNextTimeout(ctx)
			if err != nil {
				// No time left.
				return err
			}
			s.AddFuture(workflow.NewTimer(ctx, timeout), func(f workflow.Future) {
				timedOut = true
			})
		}
		s.Select(ctx)
		if timedOut {
			return errors.New("timed out")
		}
		if !receivedFirstSignal {
			a.FirstSignalTime = workflow.Now(ctx)
			receivedFirstSignal = true
		}
	}
	return nil
}
