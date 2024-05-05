package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context, name string) (string, error) {
	didUpdate := false
	l := workflow.GetLogger(ctx)
	l.Info("in wf")
	workflow.SetUpdateHandler(ctx, "update-handler", func(arg string) (string, error) {
		l.Info("in handler")
		result := "initial-result"
		if true {
			ao := workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
			}
			ctx = workflow.WithActivityOptions(ctx, ao)

			l.Info("in handler calling activity")
			err := workflow.ExecuteActivity(ctx, Activity, name+"-"+arg).Get(ctx, &result)
			if err != nil {
				return "", err
			}
		} else {
			if err := workflow.Sleep(ctx, time.Second); err != nil {
				l.Error("error while sleeping", err)
			}
		}
		l.Info("in handler did activity/sleep")
		didUpdate = true
		return result, nil
	})
	workflow.Await(ctx, func() bool { return didUpdate })
	workflow.Sleep(ctx, 5*time.Second)
	return "wf-result", nil
}

func Activity(ctx context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}

func Starter(c client.Client) {
	fmt.Println("1")
	wfID := "wf-id"
	workflowOptions := client.StartWorkflowOptions{
		ID:                    wfID,
		TaskQueue:             "hello-world",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx := context.Background()
	wfRun, err := c.ExecuteWorkflow(ctx, workflowOptions, Workflow, "wf-arg")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	fmt.Println("2")
	updHandle, err := c.UpdateWorkflow(ctx, wfID, wfRun.GetRunID(), "update-handler", "update-arg")
	if err != nil {
		log.Fatalln("error executing update", err)
	}
	fmt.Println("3")
	var updResult string
	if err := updHandle.Get(ctx, &updResult); err != nil {
		log.Fatalln("error getting update result")
	}
	fmt.Println("update result:", updResult)

	fmt.Println("4")
	var wfResult string
	err = wfRun.Get(ctx, &wfResult)
	if err != nil {
		log.Fatalln("error getting workflow result", err)
	}
	fmt.Println("workflow result:", wfResult)
}

func Worker(c client.Client) {
	w := worker.New(c, "hello-world", worker.Options{})
	w.RegisterActivity(Activity)
	w.RegisterWorkflow(Workflow)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("unable to start worker", err)
	}
}

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	go Worker(c)
	Starter(c)
}
