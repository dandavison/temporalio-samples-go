package main

import (
	"log"

	"github.com/temporalio/samples-go/pollupdatebug"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, pollupdatebug.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(pollupdatebug.Workflow)
	w.RegisterActivity(pollupdatebug.ActivityCalledByUpdate)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
