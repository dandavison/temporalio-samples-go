package pollupdatebug

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	UpdateNameBlocking = "my-blocking-update"
	SignalNameDone     = "done"
	TaskQueueName      = "pollupdatebug-tq"
)

func Workflow(ctx workflow.Context) (string, error) {
	updateFinished := false
	mayExit := false

	Must1(workflow.SetUpdateHandler(ctx, UpdateNameBlocking, func(ctx workflow.Context) (string, error) {
		activityOptions := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		})
		Must1(workflow.ExecuteActivity(activityOptions, ActivityCalledByUpdate).Get(ctx, nil))
		updateFinished = true
		return "update-result", nil
	}))

	signalChan := workflow.GetSignalChannel(ctx, SignalNameDone)
	workflow.Go(ctx, func(ctx workflow.Context) {
		signalChan.Receive(ctx, nil)
		mayExit = true
	})

	Must1(workflow.Await(ctx, func() bool {
		return updateFinished && mayExit
	}))

	return "workflow-result", nil
}

func ActivityCalledByUpdate(ctx context.Context) error {
	time.Sleep(1 * time.Second)
	return nil
}

func Must1(err error) {
	if err != nil {
		panic(err)
	}
}
