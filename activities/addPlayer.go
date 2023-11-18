package triviagame

import (
	"context"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func AddPlayerActivity(ctx context.Context, activityInput resources.AddPlayerActivityInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("GetRandomCategoryActivity")

	c, err := client.Dial(resources.GetClientOptions("workflow"))
	if err != nil {
		return err
	}
	defer c.Close()

	updateHandle, err := c.UpdateWorkflow(context.Background(), activityInput.WorkflowId, "", "AddPlayer", activityInput.Player)
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
