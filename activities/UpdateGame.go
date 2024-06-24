package triviagame

import (
	"context"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func UpdateGameActivity(ctx context.Context, gameId string, gameStatus string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("UpdateGameActivity")

	gameMap := make(map[string]string)
	gameMap["gameId"] = gameId
	gameMap["status"] = gameStatus

	c, err := client.Dial(resources.GetClientOptions("workflow"))
	if err != nil {
		return err
	}
	defer c.Close()

	updateHandle, err := c.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
		WorkflowID:   "trivia-game",
		UpdateName:   "UpdateGame",
		WaitForStage: client.WorkflowUpdateStageCompleted,
		Args:         []interface{}{gameMap},
	})

	if err != nil {
		return err
	}
	var updateResult bool
	err = updateHandle.Get(context.Background(), &updateResult)
	if err != nil {
		return err
	}

	return nil

}
