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

	// run activity to check player name against moderation api
	moderationInput := resources.ModerationInput{
		Url:  os.Getenv("MODERATION_URL"),
		Name: workflowInput.Player,
	}

	var moderationResult bool
	modErr := workflow.ExecuteActivity(ctx, activities.ModerationActivity, moderationInput).Get(ctx, &moderationResult)
	if modErr != nil {
		logger.Error("Moderation Activity failed.", "Error", modErr)
		return modErr
	}

	if moderationResult {
		return errors.New("Player name is invalid")
	}

	// Using update in local activity add player to game
	laCtx := workflow.WithLocalActivityOptions(ctx, setDefaultLocalActivityOptions())

	activityAddPlayerInput := resources.AddPlayerActivityInput{
		WorkflowId: workflowInput.GameWorkflowId,
		Player:     workflowInput.Player,
	}

	err := workflow.ExecuteLocalActivity(laCtx, activities.AddPlayerActivity, activityAddPlayerInput).Get(laCtx, nil)

	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}
