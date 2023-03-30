package triviagame

import (
	"os"
	"regexp"
	"strings"
	"time"

	activities "github.com/ktenzer/temporal-trivia/activities"
	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Trivia game workflow definition
func TriviaGameWorkflow(ctx workflow.Context, workflowInput resources.WorkflowInput) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 100 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Game Started")

	// Create objects for storing game results, progress and player scoring
	gameMap := make(map[int]resources.Result)
	scoreMap := make(map[string]int)
	var gameProgress resources.GameProgress
	gameProgress.NumberOfQuestions = workflowInput.NumberOfQuestions
	gameProgress.CurrentQuestion = 0

	// Setup query handler for gathering game details
	err := workflow.SetQueryHandler(ctx, "getDetails", func(input []byte) (map[int]resources.Result, error) {
		return gameMap, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for gameMap: " + err.Error())
		return err
	}

	// Setup query handler for gathering player score
	err = workflow.SetQueryHandler(ctx, "getScore", func(input []byte) (map[string]int, error) {
		return scoreMap, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for scoreMap: " + err.Error())
		return err
	}

	// Setup query handler for gathering game progress
	err = workflow.SetQueryHandler(ctx, "getProgress", func(input []byte) (resources.GameProgress, error) {
		return gameProgress, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for scoreMap: " + err.Error())
		return err
	}

	// set activity inputs for starting game
	activityInput := resources.ActivityInput{
		Key:               os.Getenv("CHATGPT_API_KEY"),
		Category:          workflowInput.Category,
		NumberOfQuestions: workflowInput.NumberOfQuestions,
	}

	// run activity to start game and pre-fetch trivia questions and answers
	err = workflow.ExecuteActivity(ctx, activities.TriviaQuestionActivity, activityInput).Get(ctx, &gameMap)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	var signal resources.Signal
	signalChan := workflow.GetSignalChannel(ctx, "game-signal")
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, more bool) {
		channel.Receive(ctx, &signal)
	})

	// loop through questions, start timer
	var questionCount int = 0
	for key, _ := range gameMap {
		timer := workflow.NewTimer(ctx, time.Duration(workflowInput.QuestionTimeLimit)*time.Second)

		var timerFired bool = false
		selector.AddFuture(timer, func(f workflow.Future) {
			err := f.Get(ctx, nil)
			if err == nil {
				logger.Info("Time limit for question has exceeded the limit of " + time.Duration(workflowInput.QuestionTimeLimit).String() + " seconds")
				timerFired = true
			}
		})

		// Loop through the number of players we expect to answer and break loop if question timer expires
		result := gameMap[key]
		var submissionsMap = make(map[string]resources.Submission)
		gameProgress.CurrentQuestion = questionCount

		for a := 0; a < workflowInput.NumberOfPlayers; a++ {
			// continue to next question if timer fires
			if timerFired {
				break
			}
			selector.Select(ctx)

			// handle duplicate answers from same player
			var submission resources.Submission
			if signal.Action == "Answer" && !isAnswerDuplicate(result.Submissions, signal.Player) {
				// if we don't receive valid answer mark as wrong and continue
				if !validateAnswer(signal.Answer) {
					submission.IsCorrect = false
					submission.Answer = signal.Answer
					submissionsMap[signal.Player] = submission
					continue
				}

				// ensure answer is upper case
				answerUpperCase := strings.ToUpper(signal.Answer)
				submission.Answer = answerUpperCase

				if result.Answer == submission.Answer {
					submission.IsCorrect = true

					if result.Winner == "" {
						result.Winner = signal.Player
						playerScore := scoreMap[signal.Player] + 2
						scoreMap[signal.Player] = playerScore
					} else {
						playerScore := scoreMap[signal.Player] + 1
						scoreMap[signal.Player] = playerScore
					}
				} else {
					submission.IsCorrect = false

					// add player to scoreboard if they don't exist
					_, ok := scoreMap[signal.Player]
					if !ok {
						scoreMap[signal.Player] = 0
					}
				}
				submissionsMap[signal.Player] = submission
				result.Submissions = submissionsMap
			} else {
				logger.Warn("Incorrect signal received", signal)
				a--
			}
			gameMap[key] = result
			questionCount++
		}
	}

	// sort scoreboard
	var scoreboard []resources.ScoreBoard
	err = workflow.ExecuteActivity(ctx, activities.LeaderBoardActivity, scoreMap).Get(ctx, &scoreboard)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	return nil
}

// validate answer
func validateAnswer(answer string) bool {
	re := regexp.MustCompile(`^[A-Da-d]$`)
	isMatch := re.MatchString(answer)

	return isMatch
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
