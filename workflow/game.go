package triviagame

import (
	"errors"

	activities "github.com/ktenzer/temporal-trivia/activities"
	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Trivia game workflow definition
func TriviaGameWorkflow(ctx workflow.Context, workflowInput resources.GameWorkflowInput) error {

	// Setup logger
	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Game Started")

	// Game configuration
	gameConfiguration := resources.NewGameConfigurationFromWorkflowInput(workflowInput)

	// Activity options
	ctx = workflow.WithActivityOptions(ctx, setDefaultActivityOptions())

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
	gp, err = gp.initGetProgressQuery(ctx, gameConfiguration.NumberOfQuestions)
	if err != nil {
		return err
	}

	// Add players to game using signal and wait for start game signal.
	// If game is not started before start game timeout, workflow will fail.
	var ps PlayerSignal
	isCancelled := ps.addPlayers(ctx, gameConfiguration, getPlayers)
	if isCancelled {
		return errors.New("Time limit of " + intToString(gameConfiguration.StartTimeLimit) + gameConfiguration.Category + " seconds for starting game has been exceeded!")
	}

	// Set trivia question category randomly, if category is empty
	if gameConfiguration.Category == "" {
		var category string
		laCtx := workflow.WithLocalActivityOptions(ctx, setDefaultLocalActivityOptions())
		workflow.ExecuteLocalActivity(laCtx, activities.GetRandomCategoryActivity).Get(laCtx, &category)
		gameConfiguration.Category = category
	}

	// Update game progress stage
	gp.Stage = "questions"

	// Run activity to start game and pre-fetch trivia questions and answers
	activityInput := triviaQuestionsActivityInput(gameConfiguration.Category, gameConfiguration.NumberOfQuestions)
	err = workflow.ExecuteActivity(ctx, activities.TriviaQuestionActivity, activityInput).Get(ctx, &getQuestions)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	// Run the game loop
	getQuestions, getPlayers = gp.runGame(ctx, gameConfiguration, getQuestions, getPlayers)

	// Provide sorted scoreboard results
	var scoreboard []resources.ScoreBoard
	err = workflow.ExecuteActivity(ctx, activities.LeaderBoardActivity, getPlayers).Get(ctx, &scoreboard)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	return nil
}
