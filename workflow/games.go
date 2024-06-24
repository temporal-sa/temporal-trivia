package triviagame

import (
	"time"

	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Trivia game workflow definition
func TriviaGamesWorkflow(ctx workflow.Context) error {

	// Setup logger
	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Games Started")

	// Initialize game players state machine, exporting as query
	getGames, err := initGamesQuery(ctx)
	if err != nil {
		return err
	}
	for {
		// Initialize update handler to add and validate players joining game
		err = updateGame(ctx, *getGames)
		if err != nil {
			logger.Error("Update failed.", "Error", err)
			return err
		}

		err := workflow.Sleep(ctx, time.Hour*24)
		if err != nil {
			return err
		}

		if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
			return workflow.NewContinueAsNewError(ctx, TriviaGamesWorkflow)
		}
	}
}
