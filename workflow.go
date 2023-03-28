package triviagame

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
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
	scoreboardMap := make(map[string]int)

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
		return scoreboardMap, nil
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
			Question: "Give me a different " + workflowInput.Category + " trivia question that starts with what is and has 4 possible answers A), B), C), or D)? Please provide a newline after the question. Give the correct Answer letter at the end.",
		}

		err := workflow.ExecuteActivity(ctx, TriviaQuestionActivity, activityInput).Get(ctx, &triviaQuestion)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}

		// Start timer based on question time limit of game
		timer := workflow.NewTimer(ctx, time.Duration(workflowInput.QuestionTimeLimit)*time.Second)

		logger.Info("Trivia question", "result", triviaQuestion)

		correctAnswer := parseAnswer(triviaQuestion)
		result.Answer = correctAnswer
		result.Question = parseQuestion(triviaQuestion)
		fmt.Println("HERE \n" + result.Question)
		gameMap[q] = result

		answersMap := parsePossibleAnswers(triviaQuestion)
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
				logger.Info("Time limit for question has exceeded the limit of " + time.Duration(workflowInput.QuestionTimeLimit).String() + " seconds")
				timerFired = true
			}
		})

		// Loop through the players we expect to answer and break loop if question timer expires
		var submissionsMap = make(map[string]resources.Submission)

		for a := 0; a < workflowInput.NumberOfPlayers; a++ {
			var submission resources.Submission

			// continue to next question if timer fires
			if timerFired {
				continue
			} else {
				selector.Select(ctx)
			}

			// handle duplicate answers from same player
			if signal.Action == "Answer" && !isAnswerDuplicate(gameMap[q].Submissions, signal.Player) {
				// if we don't receive valid answer mark as wrong and continue
				if !validateAnswer(signal.Answer) {
					submission.IsCorrect = false
					submission.Answer = signal.Answer
					continue
				}

				// ensure answer is upper case
				answerUpperCase := strings.ToUpper(signal.Answer)
				if result.Answer == answerUpperCase {
					submission.IsCorrect = true
					submission.Answer = signal.Answer

					if result.Winner == "" {
						result.Winner = signal.Player
						playerScore := scoreboardMap[signal.Player] + 2
						scoreboardMap[signal.Player] = playerScore
					} else {
						playerScore := scoreboardMap[signal.Player] + 1
						scoreboardMap[signal.Player] = playerScore
					}
				} else {
					submission.IsCorrect = false
					submission.Answer = signal.Answer

					// add player to scoreboard if they don't exist
					_, ok := scoreboardMap[signal.Player]
					if !ok {
						scoreboardMap[signal.Player] = 0
					}
				}
				submissionsMap[signal.Player] = submission
				result.Submissions = submissionsMap
				gameMap[q] = result
			} else {
				logger.Warn("Incorrect signal received", signal)
				a--
			}
		}
		gameMap[q] = result
		fmt.Println("GAME MAP: ", gameMap)
		fmt.Println("SCORE MAP: ", scoreboardMap)
	}

	// Output final score via activity
	var winners []string
	err = workflow.ExecuteActivity(ctx, ScoreTotalActivity, scoreboardMap).Get(ctx, &winners)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	logger.Info("The winners are...", winners)
	return nil
}

// Parse the question
func parseQuestion(question string) string {
	re := regexp.MustCompile(`\n[^\n]*$`)
	removedAnswer := re.ReplaceAllString(question, "")

	return removedAnswer
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

// Parse answer
func parseAnswer(question string) string {
	re := regexp.MustCompile(`\w+\s*Answer:? ([A-D])\)?`)

	match := re.FindStringSubmatch(question)
	if len(match) > 0 {
		return match[1]
	}

	return ""
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
