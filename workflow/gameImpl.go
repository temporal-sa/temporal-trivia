package triviagame

import (
	"strings"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func (gp *GameProgress) runGame(ctx workflow.Context, workflowInput resources.GameWorkflowInput, getQuestions *map[int]resources.Result,
	getPlayers *map[string]resources.Player) (*map[int]resources.Result, *map[string]resources.Player) {

	logger := workflow.GetLogger(ctx)

	as := AnswerSignal{}
	answerSelector := workflow.NewSelector(ctx)
	as.answerSignal(ctx, answerSelector)

	var questionCount int = 0
	keys := getSortedGameMap(*getQuestions)

	for _, key := range keys {
		gp.CurrentQuestion = questionCount + 1

		// Set game progress to answer phase
		gp.Stage = "answers"

		// Async timer for amount of time to receive answers
		timer := workflow.NewTimer(ctx, time.Duration(workflowInput.AnswerTimeLimit)*time.Second)

		var timerFired bool
		answerSelector.AddFuture(timer, func(f workflow.Future) {
			err := f.Get(ctx, nil)
			if err == nil {
				logger.Info("Time limit for question has exceeded the limit of " + time.Duration(workflowInput.AnswerTimeLimit).String() + " seconds")
				timerFired = true
			}
		})

		// Loop through the number of players we expect to answer and break loop if question timer expires
		result := (*getQuestions)[key]
		var submissionsMap = make(map[string]resources.Submission)

		for a := 0; a < workflowInput.NumberOfPlayers; a++ {
			// continue to next question if timer fires
			if timerFired {
				break
			}

			answerSelector.Select(ctx)

			// handle duplicate answers from same player
			var submission resources.Submission
			if as.Action == "Answer" && !isAnswerDuplicate(submissionsMap, as.Player) && key == as.Question {

				// ensure answer is upper case
				answerUpperCase := strings.ToUpper(as.Answer)
				submission.Answer = answerUpperCase

				if result.Answer == submission.Answer {
					submission.IsCorrect = true

					if result.Winner == "" {
						result.Winner = as.Player
						submission.IsFirst = true

						(*getPlayers)[as.Player] = resources.Player{
							Score: (*getPlayers)[as.Player].Score + 2,
						}
					} else {
						(*getPlayers)[as.Player] = resources.Player{
							Score: (*getPlayers)[as.Player].Score + 1,
						}
					}
				}
				submissionsMap[as.Player] = submission
				result.Submissions = submissionsMap
			} else {
				logger.Warn("Incorrect signal received", as)
				a--
			}

			(*getQuestions)[key] = result
		}

		// Set game progress to result phase
		gp.Stage = "result"

		// Sleep allowing time to display results
		workflow.Sleep(ctx, time.Duration(workflowInput.ResultTimeLimit)*time.Second)

		questionCount++
	}

	return getQuestions, getPlayers
}
