package triviagame

import (
	"errors"
	"os"
	"sort"
	"strings"
	"time"

	activities "github.com/ktenzer/temporal-trivia/activities"
	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Trivia game workflow definition
func TriviaGameWorkflow(ctx workflow.Context, workflowInput resources.WorkflowInput) error {

	// Set workflow defaults
	workflowInput = resources.SetDefaults(workflowInput)

	// Activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 300 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Game Started")

	// Async timer to cancel game if not started
	cancelTimer := workflow.NewTimer(ctx, time.Duration(30)*time.Second)

	// Setup player, question and progress state machines
	getPlayers := make(map[string]resources.Player)
	getQuestions := make(map[int]resources.Result)

	var gameProgress resources.GameProgress
	gameProgress.NumberOfQuestions = workflowInput.NumberOfQuestions
	gameProgress.CurrentQuestion = 0

	// Set game progress to start phase
	gameProgress.Stage = "start"

	// Setup query handler for gathering game questions
	err := workflow.SetQueryHandler(ctx, "getPlayers", func(input []byte) (map[string]resources.Player, error) {
		return getPlayers, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for getPlayers: " + err.Error())
		return err
	}

	// Setup query handler for gathering game questions
	err = workflow.SetQueryHandler(ctx, "getQuestions", func(input []byte) (map[int]resources.Result, error) {
		return getQuestions, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for getQuestions: " + err.Error())
		return err
	}

	// Setup query handler for gathering game progress
	err = workflow.SetQueryHandler(ctx, "getProgress", func(input []byte) (resources.GameProgress, error) {
		return gameProgress, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for gameProgress: " + err.Error())
		return err
	}

	// Add players to game using signal
	var addPlayerSignal resources.Signal
	addPlayerSignalChan := workflow.GetSignalChannel(ctx, "start-game-signal")
	addPlayerSelector := workflow.NewSelector(ctx)
	addPlayerSelector.AddReceive(addPlayerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &addPlayerSignal)
	})

	var cancelTimerFired bool = false
	addPlayerSelector.AddFuture(cancelTimer, func(f workflow.Future) {
		err := f.Get(ctx, nil)
		if err == nil {
			logger.Info("Time limit for starting game has been exceeded " + time.Duration(workflowInput.AnswerTimeLimit).String() + " seconds")
			cancelTimerFired = true
		}
	})

	playerCount := 0
	for {
		addPlayerSelector.Select(ctx)

		if addPlayerSignal.Action == "Player" && addPlayerSignal.Player != "" {
			getPlayers[addPlayerSignal.Player] = resources.Player{
				Id:    playerCount,
				Score: 0,
			}
		}

		playerCount++

		// Wait for start of game via signal
		if addPlayerSignal.Action == "StartGame" {
			break
		}

		if cancelTimerFired {
			return errors.New("Time limit for starting game has been exceeded!")
		}
	}

	// set activity inputs for starting game
	activityInput := resources.ActivityInput{
		Key:               os.Getenv("CHATGPT_API_KEY"),
		Category:          workflowInput.Category,
		NumberOfQuestions: workflowInput.NumberOfQuestions,
	}

	// run activity to start game and pre-fetch trivia questions and answers
	err = workflow.ExecuteActivity(ctx, activities.TriviaQuestionActivity, activityInput).Get(ctx, &getQuestions)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	var answerSignal resources.Signal
	answerSignalChan := workflow.GetSignalChannel(ctx, "answer-signal")
	answerSelector := workflow.NewSelector(ctx)
	answerSelector.AddReceive(answerSignalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &answerSignal)
	})

	// loop through questions, start timer
	var questionCount int = 0
	keys := getSortedGameMap(getQuestions)
	for _, key := range keys {
		gameProgress.CurrentQuestion = questionCount + 1

		// Set game progress to answer phase
		gameProgress.Stage = "answer"

		// Async timer for amount of time to receive answers
		timer := workflow.NewTimer(ctx, time.Duration(workflowInput.AnswerTimeLimit)*time.Second)

		var timerFired bool = false
		answerSelector.AddFuture(timer, func(f workflow.Future) {
			err := f.Get(ctx, nil)
			if err == nil {
				logger.Info("Time limit for question has exceeded the limit of " + time.Duration(workflowInput.AnswerTimeLimit).String() + " seconds")
				answerSignal = resources.Signal{}
				timerFired = true
			}
		})

		// Loop through the number of players we expect to answer and break loop if question timer expires
		result := getQuestions[key]
		var submissionsMap = make(map[string]resources.Submission)

		for a := 0; a < workflowInput.NumberOfPlayers; a++ {
			// continue to next question if timer fires
			if timerFired {
				break
			}

			// Update screen to question
			//getScreen = "answer"

			answerSelector.Select(ctx)

			// handle duplicate answers from same player
			var submission resources.Submission
			if answerSignal.Action == "Answer" && !isAnswerDuplicate(result.Submissions, answerSignal.Player) {

				// ensure answer is upper case
				answerUpperCase := strings.ToUpper(answerSignal.Answer)
				submission.Answer = answerUpperCase
				submission.PlayerId = getPlayers[answerSignal.Player].Id

				if result.Answer == submission.Answer {
					submission.IsCorrect = true

					if result.Winner == "" {
						result.Winner = answerSignal.Player
						submission.IsFirst = true
						//playerScore := scoreMap[answerSignal.Player] + 2
						//scoreMap[answerSignal.Player] = playerScore
						getPlayers[answerSignal.Player] = resources.Player{
							Id:    getPlayers[answerSignal.Player].Id,
							Score: getPlayers[answerSignal.Player].Score + 2,
						}
					} else {
						//playerScore := scoreMap[answerSignal.Player] + 1
						//scoreMap[answerSignal.Player] = playerScore
						getPlayers[answerSignal.Player] = resources.Player{
							Id:    getPlayers[answerSignal.Player].Id,
							Score: getPlayers[answerSignal.Player].Score + 1,
						}
					}
				} else {
					submission.IsCorrect = false

					// add player to scoreboard if they don't exist
					//_, ok := scoreMap[answerSignal.Player]
					//if !ok {
					//	scoreMap[answerSignal.Player] = 0
					//}
				}
				submissionsMap[answerSignal.Player] = submission
				result.Submissions = submissionsMap
			} else {
				logger.Warn("Incorrect signal received", answerSignal)
				a--
			}

			getQuestions[key] = result
		}

		// Set game progress to result phase
		gameProgress.Stage = "result"

		// Sleep allowing time to display results
		workflow.Sleep(ctx, time.Duration(workflowInput.ResultTimeLimit)*time.Second)

		questionCount++
	}

	// sort scoreboard
	var scoreboard []resources.ScoreBoard
	err = workflow.ExecuteActivity(ctx, activities.LeaderBoardActivity, getPlayers).Get(ctx, &scoreboard)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	return nil
}

// Detect answer duplication
func isAnswerDuplicate(submissions map[string]resources.Submission, player string) bool {
	for submittedPlayer, _ := range submissions {
		if player == submittedPlayer {
			return true
		}
	}
	return false
}

// Sort gameMap
func getSortedGameMap(gameMap map[int]resources.Result) []int {

	keys := make([]int, 0, len(gameMap))
	for k := range gameMap {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	return keys
}
