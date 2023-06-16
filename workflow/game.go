package triviagame

import (
	"errors"
	"os"
	"strings"
	"time"

	activities "github.com/ktenzer/temporal-trivia/activities"
	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Trivia game workflow definition
func TriviaGameWorkflow(ctx workflow.Context, workflowInput resources.GameWorkflowInput) error {

	// Set workflow defaults
	workflowInput = resources.SetDefaults(workflowInput)

	// Activity options
	ctx = workflow.WithActivityOptions(ctx, setDefaultActivityOptions())

	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Game Started")

	// Async timer to cancel game if not started
	cancelTimer := workflow.NewTimer(ctx, time.Duration(workflowInput.StartTimeLimit)*time.Second)

	// Initialize game players state machine, exporting as query
	getPlayers, err := initGetPlayersQuery(ctx)
	if err != nil {
		return err
	}

	// Initialize game questions state machine, exporting as query
	getQuestions, err := initGetQuestionsQuery(ctx)
	if err != nil {
		return err
	}

	// Initialize game progress state machine, exporting as query
	var gp GameProgress
	gp, err = gp.initGetProgressQuery(ctx, workflowInput.NumberOfQuestions)
	if err != nil {
		return err
	}

	// Add players to game using signal and wait for start game signal
	var gs GameSignal
	addPlayerSelector := gs.gameSignal(ctx)

	var cancelTimerFired bool = false
	addPlayerSelector.AddFuture(cancelTimer, func(f workflow.Future) {
		err := f.Get(ctx, nil)
		if err == nil {
			logger.Info("Time limit for starting game has been exceeded " + time.Duration(workflowInput.StartTimeLimit).String() + " seconds")
			cancelTimerFired = true
		}
	})

	// Add players to game
	for {
		addPlayerSelector.Select(ctx)

		if gs.Action == "Player" && gs.Player != "" {
			if _, ok := getPlayers[gs.Player]; ok {
				getPlayers[gs.Player] = resources.Player{
					Score: 0,
				}
			}
		}

		// Wait for start of game via signal
		if gs.Action == "StartGame" {
			break
		}

		if cancelTimerFired {
			return errors.New("Time limit for starting game has been exceeded!")
		}
	}

	// Set game progress to generation questions phase
	gp.Stage = "questions"

	// run activity to start game and pre-fetch trivia questions and answers
	activityInput := setTriviaQuestionsActivityInput(os.Getenv("CHATGPT_API_KEY"), workflowInput.Category, workflowInput.NumberOfQuestions)
	err = workflow.ExecuteActivity(ctx, activities.TriviaQuestionActivity, activityInput).Get(ctx, &getQuestions)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	// Recieve player answer via signals
	var as GameSignal
	answerSelector := as.answerSignal(ctx)

	// loop through questions, start timer
	var questionCount int = 0
	keys := getSortedGameMap(getQuestions)
	for _, key := range keys {
		gp.CurrentQuestion = questionCount + 1

		// Set game progress to answer phase
		gp.Stage = "answers"

		// Async timer for amount of time to receive answers
		timer := workflow.NewTimer(ctx, time.Duration(workflowInput.AnswerTimeLimit)*time.Second)

		var timerFired bool = false
		answerSelector.AddFuture(timer, func(f workflow.Future) {
			err := f.Get(ctx, nil)
			if err == nil {
				logger.Info("Time limit for question has exceeded the limit of " + time.Duration(workflowInput.AnswerTimeLimit).String() + " seconds")
				as = GameSignal{}
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

			answerSelector.Select(ctx)

			// handle duplicate answers from same player
			var submission resources.Submission
			if as.Action == "Answer" && !isAnswerDuplicate(result.Submissions, as.Player) {

				// ensure answer is upper case
				answerUpperCase := strings.ToUpper(as.Answer)
				submission.Answer = answerUpperCase

				if result.Answer == submission.Answer {
					submission.IsCorrect = true

					if result.Winner == "" {
						result.Winner = as.Player
						submission.IsFirst = true

						getPlayers[as.Player] = resources.Player{
							Score: getPlayers[as.Player].Score + 2,
						}
					} else {
						getPlayers[as.Player] = resources.Player{
							Score: getPlayers[as.Player].Score + 1,
						}
					}
				} else {
					submission.IsCorrect = false
				}
				submissionsMap[as.Player] = submission
				result.Submissions = submissionsMap
			} else {
				logger.Warn("Incorrect signal received", as)
				a--
			}

			getQuestions[key] = result
		}

		// Set game progress to result phase
		gp.Stage = "result"

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
