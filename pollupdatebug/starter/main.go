package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"github.com/temporalio/samples-go/pollupdatebug"
	"go.temporal.io/sdk/client"
)

func mainWithUWS() {
	c := Must(client.Dial(client.Options{}))
	defer c.Close()

	ctx := context.Background()

	updOpts := client.UpdateWorkflowOptions{
		UpdateName:   pollupdatebug.UpdateNameBlocking,
		WaitForStage: client.WorkflowUpdateStageAccepted,
	}
	updOp := client.NewUpdateWithStartWorkflowOperation(updOpts)

	Must(c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                 "pollupdatebug-workflow-ID-" + uuid.New(),
		TaskQueue:          pollupdatebug.TaskQueueName,
		WithStartOperation: updOp,
	}, pollupdatebug.Workflow))

	// The first PollUpdate works as expected.
	updHandle1 := Must(updOp.Get(ctx))
	var result1 string
	Must1(updHandle1.Get(ctx, &result1))
	if result1 != "update-result" {
		log.Fatalln("Unexpected result from first update:", result1)
	}

	// But the second PollUpdate fails.
	updHandle2 := Must(updOp.Get(ctx))
	var result2 string
	if err := updHandle2.Get(ctx, &result2); err != nil {
		// By this point, in the server, completionCallback has been called and deleted the update
		// from the registry, and it is not found in mutable state. But why -- the second
		// PollUpdate call should work shouldn't it?
		log.Fatalln("Error obtaining second update result:", err)
	}
	if result2 != "update-result" {
		log.Fatalln("Unexpected result from second update:", result2)
	}
}

func mainWithoutUWS() {
	// This doesn't use UWS. It works fine; it's just here for comparison.
	c := Must(client.Dial(client.Options{}))
	defer c.Close()

	ctx := context.Background()

	updOpts := client.UpdateWorkflowOptions{
		UpdateName:   pollupdatebug.UpdateNameBlocking,
		WaitForStage: client.WorkflowUpdateStageAccepted,
	}

	wfRun := Must(c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "pollupdatebug-workflow-ID-" + uuid.New(),
		TaskQueue: pollupdatebug.TaskQueueName,
	}, pollupdatebug.Workflow))

	updOpts.WorkflowID = wfRun.GetID()
	updHandle1 := Must(c.UpdateWorkflow(ctx, updOpts))

	var result1 string
	Must1(updHandle1.Get(ctx, &result1))
	if result1 != "update-result" {
		log.Fatalln("Unexpected result from first update:", result1)
	}

	var result2 string
	if err := updHandle1.Get(ctx, &result2); err != nil {
		log.Fatalln("Error obtaining second update result:", err)
	}
	if result2 != "update-result" {
		log.Fatalln("Unexpected result from second update:", result2)
	}
}

func main() {
	uws := true
	if uws {
		mainWithUWS()
	} else {
		mainWithoutUWS()
	}
}

func Must1(err error) {
	if err != nil {
		panic(err)
	}
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
