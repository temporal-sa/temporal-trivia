package main

import (
	"log"

	activities "github.com/ktenzer/temporal-trivia/activities"
	"github.com/ktenzer/temporal-trivia/resources"
	workflow "github.com/ktenzer/temporal-trivia/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(resources.GetClientOptions("worker"))
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "trivia-game", worker.Options{})

	w.RegisterWorkflow(workflow.TriviaGameWorkflow)
	w.RegisterActivity(activities.TriviaQuestionActivity)
	w.RegisterActivity(activities.LeaderBoardActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
