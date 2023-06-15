package triviagame

import (
	"errors"

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
	ctx = workflow.WithLocalActivityOptions(ctx, setDefaultLocalActivityOptions())

	activityInput := resources.QueryPlayerActivityInput{
		WorkflowId: workflowInput.GameWorkflowId,
		Player:     workflowInput.Player,
		QueryType:  "normal",
	}

	// run activity to start game and pre-fetch trivia questions and answers
	isPlayer := false
	err := workflow.ExecuteLocalActivity(ctx, activities.QueryPlayerActivity, activityInput).Get(ctx, &isPlayer)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	if isPlayer {
		return errors.New("Player " + workflowInput.Player + " already exists!")
	}

	// TODO: Add Activity to check player name is language compliant

	// Add player via signal
	addPlayerSignal := resources.Signal{
		Action: "Player",
		Player: workflowInput.Player,
	}

	err = workflow.SignalExternalWorkflow(ctx, workflowInput.GameWorkflowId, "", "start-game-signal", addPlayerSignal).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure player is added
	activityInput = resources.QueryPlayerActivityInput{
		WorkflowId: workflowInput.GameWorkflowId,
		Player:     workflowInput.Player,
		QueryType:  "poll",
	}

	err = workflow.ExecuteLocalActivity(ctx, activities.QueryPlayerActivity, activityInput).Get(ctx, &isPlayer)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	if !isPlayer {
		return errors.New("Player " + workflowInput.Player + " could not be added to the game")
	}

	return nil
}
