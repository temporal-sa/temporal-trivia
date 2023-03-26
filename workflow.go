package triviagame

import (
	"fmt"
	"regexp"
	"time"

	"github.com/ktenzer/triviagame/resources"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Trivia game workflow definition
func Workflow(ctx workflow.Context, workflowInput resources.WorkflowInput) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Game Started")

	// Create maps for storing game results and user scoring
	gameMap := make(map[int]resources.Result)
	scoreMap := make(map[string]int)

	// Setup query handler for tracking game progress
	err := workflow.SetQueryHandler(ctx, "getDetails", func(input []byte) (map[int]resources.Result, error) {
		return gameMap, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for gameMap: " + err.Error())
		return err
	}

	// Setup query handler for tracking game progress
	err = workflow.SetQueryHandler(ctx, "getScore", func(input []byte) (map[string]int, error) {
		return scoreMap, nil
	})

	if err != nil {
		logger.Error("SetQueryHandler failed for scoreMap: " + err.Error())
		return err
	}

	// Loop through the number of questions
	for q := 0; q < workflowInput.NumberOfQuestions; q++ {

		var result resources.Result
		var triviaQuestion string

		activityInput := resources.ActivityInput{
			Key:      workflowInput.Key,
			Question: "Give me a " + workflowInput.Category + " trivia question that starts with what is?",
		}

		err := workflow.ExecuteActivity(ctx, TriviaQuestionActivity, activityInput).Get(ctx, &triviaQuestion)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}

		result.Question = triviaQuestion
		logger.Info("Trivia question", "result", triviaQuestion)
		gameMap[q] = result

		var signal resources.Signal
		signalChan := workflow.GetSignalChannel(ctx, "game-signal")
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, more bool) {
			channel.Receive(ctx, &signal)
		})

		for a := 0; a < workflowInput.NumberOfPlayers; a++ {
			selector.Select(ctx)

			if signal.Action == "Answer" {
				question := parseQuestion(triviaQuestion)
				activityInput := resources.ActivityInput{
					Key:      workflowInput.Key,
					Question: "Is " + signal.Answer + " " + question + " Please answer true or false followed by explanation.",
				}

				var triviaAnswer string
				err := workflow.ExecuteActivity(ctx, TriviaQuestionActivity, activityInput).Get(ctx, &triviaAnswer)
				if err != nil {
					logger.Error("Activity failed.", "Error", err)
					return err
				}

				logger.Info("Trivia answer from user", signal.User, triviaAnswer)

				answerResult, answerDetails := parseAnswer(triviaAnswer)
				result.AnswerDetails = answerDetails
				isCorrect := validateAnswer(answerResult)

				if isCorrect {
					result.CorrectAnswers = append(result.CorrectAnswers, signal.User)
					if result.Winner == "" {
						result.Winner = signal.User
						scoreMap[signal.User] = scoreMap[signal.User] + 1
					}
				} else {
					result.WrongAnswers = append(result.WrongAnswers, signal.User)
				}
				gameMap[q] = result
			}
		}
		gameMap[q] = result
		fmt.Println("GAME MAP: ", gameMap)
		fmt.Println("SCORE MAP: ", scoreMap)
	}

	// Output final score via activity
	var scoreOutput resources.ActivityScoreOutput
	err = workflow.ExecuteActivity(ctx, ScoreTotalActivity, scoreMap).Get(ctx, &scoreOutput)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	logger.Info("The winners are...", scoreOutput)
	return nil
}

func parseQuestion(question string) string {
	re := regexp.MustCompile(`What\s+is\s+(.*)`)
	match := re.FindStringSubmatch(question)

	if len(match) != 0 {
		return match[1]
	}

	return ""
}

func parseAnswer(answer string) (string, string) {
	re := regexp.MustCompile(`(True|False)(.*)`)
	match := re.FindStringSubmatch(answer)

	if len(match) != 0 {
		return match[1], match[2]
	}

	return "", ""
}

func validateAnswer(answerResult string) bool {
	if answerResult == "True" {
		return true
	} else {
		return false
	}
}
