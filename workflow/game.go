package triviagame

import (

	//"strings"

	"errors"

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
	//var gs GameSignal
	var ps PlayerSignal
	isCancelled := ps.runPlayerLogic(ctx, workflowInput, getPlayers)
	if isCancelled {
		return errors.New("Time limit for starting game has been exceeded!")
	}

	// Set Category if not provided
	if workflowInput.Category == "" {
		var category string
		laCtx := workflow.WithLocalActivityOptions(ctx, setDefaultLocalActivityOptions())
		workflow.ExecuteLocalActivity(laCtx, activities.GetRandomCategoryActivity).Get(laCtx, &category)
		workflowInput.Category = category
	}

	// Set game progress to generation questions phase
	gp.Stage = "questions"

	// run activity to start game and pre-fetch trivia questions and answers
	activityInput := triviaQuestionsActivityInput(workflowInput.Category, workflowInput.NumberOfQuestions)
	err = workflow.ExecuteActivity(ctx, activities.TriviaQuestionActivity, activityInput).Get(ctx, &getQuestions)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	getQuestions, getPlayers = gp.runGameLogic(ctx, workflowInput, getQuestions, getPlayers)

	// sort scoreboard
	var scoreboard []resources.ScoreBoard
	err = workflow.ExecuteActivity(ctx, activities.LeaderBoardActivity, getPlayers).Get(ctx, &scoreboard)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	return nil
}
