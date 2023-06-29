package triviagame

import (
	"errors"
	"os"

	activities "github.com/ktenzer/temporal-trivia/activities"
	"github.com/ktenzer/temporal-trivia/resources"
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
	"go.temporal.io/sdk/workflow"
)

// Add player workflow definition
func AddPlayerWorkflow(ctx workflow.Context, workflowInput resources.AddPlayerWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Game Started")

	// Activity options
	ctx = workflow.WithActivityOptions(ctx, setDefaultActivityOptions())
	laCtx := workflow.WithLocalActivityOptions(ctx, setDefaultLocalActivityOptions())

	activityInput := resources.QueryPlayerActivityInput{
		WorkflowId: workflowInput.GameWorkflowId,
		Player:     workflowInput.Player,
		QueryType:  "normal",
	}

	// run activity to start game and pre-fetch trivia questions and answers
	var isPlayer bool
	err := workflow.ExecuteLocalActivity(laCtx, activities.QueryPlayerActivity, activityInput).Get(laCtx, &isPlayer)
	if err != nil {
		logger.Error("Query Player Activity failed.", "Error", err)
		return err
	}

	if isPlayer {
		return errors.New("Player " + workflowInput.Player + " already exists!")
	}

	// run activity to check for valid name
	moderationInput := resources.ModerationInput{
		Url:  os.Getenv("MODERATION_URL"),
		Name: workflowInput.Player,
	}

	// run moderation activity to check provided name
	var moderationResult bool
	modErr := workflow.ExecuteActivity(ctx, activities.ModerationActivity, moderationInput).Get(ctx, &moderationResult)
	if modErr != nil {
		logger.Error("Moderation Activity failed.", "Error", modErr)
		return modErr
	}

	if moderationResult {
		return errors.New("Player name is invalid")
	}

	// Add player via signal
	addPlayerSignal := PlayerSignal{
		Action: "Player",
		Player: workflowInput.Player,
	}

	err = workflow.SignalExternalWorkflow(ctx, workflowInput.GameWorkflowId, "", "add-player-signal", addPlayerSignal).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure player is added
	activityInput = resources.QueryPlayerActivityInput{
		WorkflowId: workflowInput.GameWorkflowId,
		Player:     workflowInput.Player,
		QueryType:  "poll",
	}

	err = workflow.ExecuteLocalActivity(laCtx, activities.QueryPlayerActivity, activityInput).Get(laCtx, &isPlayer)
	if err != nil {
		logger.Error("Query Player Activity failed before adding to game.", "Error", err)
		return err
	}

	if !isPlayer {
		return errors.New("Player " + workflowInput.Player + " could not be added to the game")
	}

	return nil
}
