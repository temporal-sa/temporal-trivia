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

	// Initialize update handler to add and validate players joining game
	err = updatePlayer(ctx, *getPlayers)
	if err != nil {
		logger.Error("Update failed.", "Error", err)
		return err
	}

	laCtx := workflow.WithLocalActivityOptions(ctx, setDefaultLocalActivityOptions())
	err = workflow.ExecuteLocalActivity(laCtx, activities.UpdateGameActivity, workflowInput.GameId, "starting").Get(laCtx, nil)

	if err != nil {
		return errors.New(err.Error())
	}

	// Start timer and wait for timer to fire or start game signal
	isCancelled := addPlayers(ctx, gameConfiguration, getPlayers)
	if isCancelled {
		// Update game state to be completed
		err = workflow.ExecuteLocalActivity(laCtx, activities.UpdateGameActivity, workflowInput.GameId, "completed").Get(laCtx, nil)

		if err != nil {
			return errors.New(err.Error() + " Time limit of " + intToString(gameConfiguration.StartTimeLimit) + gameConfiguration.Category + " seconds for starting game has been exceeded!")
		}

		return errors.New("Time limit of " + intToString(gameConfiguration.StartTimeLimit) + gameConfiguration.Category + " seconds for starting game has been exceeded!")
	}

	// Using game state to be started
	err = workflow.ExecuteLocalActivity(laCtx, activities.UpdateGameActivity, workflowInput.GameId, "running").Get(laCtx, nil)

	if err != nil {
		return errors.New(err.Error())
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

	logger.Info(gameConfiguration.Category)
	if gameConfiguration.Category == "temporal" {
		err = workflow.ExecuteActivity(ctx, activities.TriviaQuestionKapaAI, activityInput).Get(ctx, &getQuestions)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}
	} else {
		err = workflow.ExecuteActivity(ctx, activities.TriviaQuestionChatGPT, activityInput).Get(ctx, &getQuestions)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}
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

	// Update game state to be completed
	err = workflow.ExecuteLocalActivity(laCtx, activities.UpdateGameActivity, workflowInput.GameId, "completed").Get(laCtx, nil)

	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}
