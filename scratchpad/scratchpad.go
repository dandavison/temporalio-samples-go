package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}

func Starter(c client.Client) {
	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "hello-world",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	fmt.Printf("%v\n", result)
}

func Worker(c client.Client) {
	w := worker.New(c, "hello-world", worker.Options{})
	w.RegisterWorkflow(Workflow)
	w.RegisterActivity(Activity)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func main() {
	c, err := client.Dial(client.Options{Logger: noopLogger{}})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	go Worker(c)
	Starter(c)
}

type noopLogger struct{}

func (l noopLogger) Debug(string, ...interface{}) {}
func (l noopLogger) Info(string, ...interface{})  {}
func (l noopLogger) Warn(string, ...interface{})  {}
func (l noopLogger) Error(string, ...interface{}) {}
