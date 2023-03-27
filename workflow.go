package triviagame

import (
	"fmt"
	"regexp"
	"strings"
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

		// Start timer based on question time limit of game
		timer := workflow.NewTimer(ctx, workflowInput.QuestionTimeLimit)

		activityInput := resources.ActivityInput{
			Key:      workflowInput.Key,
			Question: "Give me a different " + workflowInput.Category + " trivia question that starts with what is and has 4 possible answers A)-D)? Please provide a newline after the question. Do not provide answer.",
		}

		err := workflow.ExecuteActivity(ctx, TriviaQuestionActivity, activityInput).Get(ctx, &triviaQuestion)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}

		result.Question = triviaQuestion
		logger.Info("Trivia question", "result", triviaQuestion)
		gameMap[q] = result

		answersMap := parsePossibleAnswers(triviaQuestion)
		fmt.Println("HERE")
		fmt.Println(answersMap)
		result.MultipleChoiceMap = answersMap

		var signal resources.Signal
		signalChan := workflow.GetSignalChannel(ctx, "game-signal")
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, more bool) {
			channel.Receive(ctx, &signal)
		})

		var timerFired bool = false
		selector.AddFuture(timer, func(f workflow.Future) {
			err := f.Get(ctx, nil)
			if err == nil {
				logger.Info("Time limit for question has exceeded the limit of " + workflowInput.QuestionTimeLimit.String() + " seconds")
				timerFired = true
			}
		})

		// Loop through the players we expect to answer and break loop if question timer expires
		for a := 0; a < workflowInput.NumberOfPlayers; a++ {
			if timerFired {
				continue
			} else {
				selector.Select(ctx)
			}

			// handle duplicate answers from same player
			if signal.Action == "Answer" && !isAnswerDuplicate(gameMap[q].CorrectAnswers, signal.User) &&
				!isAnswerDuplicate(gameMap[q].WrongAnswers, signal.User) {

				// if we don't receive valid answer mark as wrong and continue
				if !validateAnswer(signal.Answer) {
					result.WrongAnswers = append(result.WrongAnswers, signal.User)
					continue
				}

				// ensure answer is upper case
				answerUpperCase := strings.ToUpper(signal.Answer)

				question := parseQuestion(triviaQuestion)
				activityInput := resources.ActivityInput{
					Key:      workflowInput.Key,
					Question: "Is " + answersMap[answerUpperCase] + " " + question + " Please answer true or false followed by explanation.",
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
				isCorrect := validateAnswerResult(answerResult)

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
			} else {
				logger.Warn("Incorrect signal received", signal)
				a--
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

// Parse the question
func parseQuestion(question string) string {
	re := regexp.MustCompile(`^What is (.+)\?`)
	match := re.FindStringSubmatch(question)

	if len(match) == 2 {
		return match[1]
	}

	return ""
}

// Parse the possible answers
func parsePossibleAnswers(question string) map[string]string {
	re := regexp.MustCompile(`([A-Z])\) (\w+(?: \w+)*)`)
	matches := re.FindAllStringSubmatch(question, -1)

	answers := make(map[string]string)
	for _, match := range matches {
		answers[match[1]] = match[2]
	}

	return answers
}

// Parse response of answer to determine if correct
func parseAnswer(answer string) (string, string) {
	re := regexp.MustCompile(`(True|False)(.*)`)
	match := re.FindStringSubmatch(answer)

	if len(match) != 0 {
		return match[1], match[2]
	}

	return "", ""
}

// validate answer
func validateAnswer(answer string) bool {
	re := regexp.MustCompile(`^[A-Da-d]$`)
	isMatch := re.MatchString(answer)

	return isMatch
}

// Validation for answer result
func validateAnswerResult(answerResult string) bool {
	if answerResult == "True" {
		return true
	} else {
		return false
	}
}

// Detect answer duplication
func isAnswerDuplicate(list []string, str string) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == str {
			return true
		}
	}
	return false
}
