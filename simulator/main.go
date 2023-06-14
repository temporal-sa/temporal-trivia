package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	triviagame "github.com/ktenzer/temporal-trivia/workflow"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(resources.GetClientOptions("workflow"))
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowId := "Temporal_Trivia_Test"

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowId,
		TaskQueue: "trivia-game",
	}

	// Set input using defaults
	input := resources.WorkflowInput{}
	input = resources.SetDefaults(input)

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, triviagame.TriviaGameWorkflow, input)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// loop through player list and add to game
	for p := 0; p < input.NumberOfPlayers; p++ {
		addPlayerSignal := resources.Signal{
			Action: "Player",
			Player: "player" + strconv.Itoa(p),
		}

		err = Signal(c, addPlayerSignal, workflowId, "start-game-signal")
		if err != nil {
			log.Fatalln("Error sending the Signal", err)
		}
	}

	// start game
	startGameSignal := resources.Signal{
		Action: "StartGame",
	}

	err = Signal(c, startGameSignal, workflowId, "start-game-signal")
	if err != nil {
		log.Fatalln("Error sending the Signal", err)
	}

	// loop through number of questions and check with game stage to provide answers
	for i := 0; i < input.NumberOfQuestions; i++ {
		log.Println("Game is on question " + strconv.Itoa(i) + " of " + strconv.Itoa(input.NumberOfQuestions))

		for {
			gameProgress, err := getGameProgress(c, workflowId)
			if err != nil {
				log.Fatalln("Error sending the Query", err)
			}

			log.Println("Game stage is " + gameProgress.Stage)

			if gameProgress.CurrentQuestion == i+1 && gameProgress.Stage == "answers" {
				break
			}
			time.Sleep(1 * time.Second)
		}

		for p := 0; p < input.NumberOfPlayers; p++ {
			setRandomSeed()
			randomLetter := getRandomLetter()

			log.Println("Player player" + strconv.Itoa(p) + " answer is " + randomLetter)
			answerSignal := resources.Signal{
				Action: "Answer",
				Player: "player" + strconv.Itoa(p),
				Answer: randomLetter,
			}

			err = Signal(c, answerSignal, workflowId, "answer-signal")
			if err != nil {
				log.Fatalln("Error sending the Signal", err)
			}

			time.Sleep(1 * time.Second)
		}
	}

}

func Signal(c client.Client, signal resources.Signal, workflowId string, signalType string) error {

	err := c.SignalWorkflow(context.Background(), workflowId, "", signalType, signal)
	if err != nil {
		return err
	}

	log.Println("Workflow[" + workflowId + "] Signaled")

	return nil
}

func getQuestions(c client.Client, workflowId string) (map[int]resources.Result, error) {
	resp, err := c.QueryWorkflow(context.Background(), workflowId, "", "getQuestions")
	if err != nil {
		return nil, err
	}

	var result map[int]resources.Result
	if err := resp.Get(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func getGameProgress(c client.Client, workflowId string) (resources.GameProgress, error) {
	resp, err := c.QueryWorkflow(context.Background(), workflowId, "", "getProgress")
	var result resources.GameProgress

	if err != nil {
		return result, err
	}

	if err := resp.Get(&result); err != nil {
		return result, err
	}

	return result, nil
}

func setRandomSeed() {
	rand.Seed(time.Now().UnixNano())
}

func getRandomLetter() string {
	letters := []rune{'a', 'b', 'c', 'd'}
	randomIndex := rand.Intn(len(letters))
	return string(letters[randomIndex])
}
