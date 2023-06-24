package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	. "github.com/ktenzer/temporal-trivia/resources"
	triviagame "github.com/ktenzer/temporal-trivia/workflow"
	"go.temporal.io/sdk/client"
)

var gameWorkflowId = "Temporal_Trivia_Simulator_Game_" + strconv.Itoa(12345)

func main() {
	c, err := client.Dial(resources.GetClientOptions("workflow"))
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	gameWorkflowOptions := client.StartWorkflowOptions{
		ID:        gameWorkflowId,
		TaskQueue: "trivia-game",
	}

	// Set input using defaults
	gameWorkflowInput := resources.GameWorkflowInput{}
	gameWorkflowInput.NumberOfPlayers = 2
	gameWorkflowInput.NumberOfQuestions = 5

	gameWorkflow, err := c.ExecuteWorkflow(context.Background(), gameWorkflowOptions, triviagame.TriviaGameWorkflow, gameWorkflowInput)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started game workflow", "WorkflowID", gameWorkflow.GetID(), "RunID", gameWorkflow.GetRunID())

	// loop through player list and add to game
	time.Sleep(1 * time.Second)
	for p := 0; p < gameWorkflowInput.NumberOfPlayers; p++ {
		player := "player" + strconv.Itoa(p)
		var playerWorkflowId = "Temporal_Trivia_Simulator_Player_" + player

		playerWorkflowOptions := client.StartWorkflowOptions{
			ID:        playerWorkflowId,
			TaskQueue: "trivia-game",
		}

		playerWorkflowInput := resources.AddPlayerWorkflowInput{
			GameWorkflowId: gameWorkflowId,
			Player:         player,
		}

		addPlayerWorkflow, err := c.ExecuteWorkflow(context.Background(), playerWorkflowOptions, triviagame.AddPlayerWorkflow, playerWorkflowInput)
		if err != nil {
			log.Fatalln("Unable to execute workflow", err)
		}

		log.Println("Started player workflow for player "+player, "WorkflowID", addPlayerWorkflow.GetID(), "RunID", addPlayerWorkflow.GetRunID())

		// synchronously wait for add player workflow to complete
		err = addPlayerWorkflow.Get(context.Background(), nil)
		if err != nil {
			log.Fatalln("Unable get workflow result", err)
		}
	}

	// start game
	startGameSignal := triviagame.GameSignal{
		Action: "StartGame",
	}

	err = GameSignal(c, startGameSignal, gameWorkflowId, "start-game-signal")
	if err != nil {
		log.Fatalln("Error sending the Signal", err)
	}

	// loop through number of questions and check with game stage to provide answers
	for i := 0; i < gameWorkflowInput.NumberOfQuestions; i++ {
		questionNumber := i + 1
		log.Println("Game is on question " + strconv.Itoa(questionNumber) + " of " + strconv.Itoa(gameWorkflowInput.NumberOfQuestions))

		for {
			gameProgress, err := getGameProgress(c, gameWorkflowId)
			if err != nil {
				log.Fatalln("Error sending the Query", err)
			}

			log.Println("Game stage is " + gameProgress.Stage)

			if gameProgress.CurrentQuestion == questionNumber && gameProgress.Stage == "answers" {
				break
			}
			time.Sleep(1 * time.Second)
		}

		for p := 0; p < gameWorkflowInput.NumberOfPlayers; p++ {
			setRandomSeed()
			randomLetter := getRandomLetter()

			log.Println("Player player" + strconv.Itoa(p) + " answer is " + randomLetter)
			answerSignal := triviagame.AnswerSignal{
				Action:   "Answer",
				Player:   "player" + strconv.Itoa(p),
				Question: questionNumber,
				Answer:   randomLetter,
			}

			err = AnswerSignal(c, answerSignal, gameWorkflowId, AnswerSignalChannelName)
			if err != nil {
				log.Fatalln("Error sending the Signal", err)
			}

			time.Sleep(1 * time.Second)
		}
	}

}

func GameSignal(c client.Client, signal triviagame.GameSignal, workflowId string, signalType string) error {

	err := c.SignalWorkflow(context.Background(), workflowId, "", signalType, signal)
	if err != nil {
		return err
	}

	log.Println("Workflow[" + workflowId + "] Signaled")

	return nil
}

func AnswerSignal(c client.Client, signal triviagame.AnswerSignal, workflowId string, signalType string) error {

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

func getGameProgress(c client.Client, workflowId string) (triviagame.GameProgress, error) {
	resp, err := c.QueryWorkflow(context.Background(), workflowId, "", "getProgress")
	var result triviagame.GameProgress

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
