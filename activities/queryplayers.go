package triviagame

import (
	"context"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	"go.temporal.io/sdk/client"
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func QueryPlayerActivity(ctx context.Context, activityInput resources.QueryPlayerActivityInput) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("QueryPlayerActivity")

	c, err := client.Dial(resources.GetClientOptions("workflow"))
	if err != nil {
		logger.Error("Unable to create client", err)
		return false, err
	}
	defer c.Close()

	var isPlayer bool
	if activityInput.QueryType == "poll" {
		for {
			resp, err := c.QueryWorkflow(context.Background(), activityInput.WorkflowId, "", "getPlayers")
			var getPlayers map[string]resources.Player

			if err != nil {
				return isPlayer, err
			}

			if err := resp.Get(&getPlayers); err != nil {
				return isPlayer, err
			}

			if _, ok := getPlayers[activityInput.Player]; ok {
				isPlayer = true
				break
			}

			time.Sleep(1 * time.Second)
		}
	} else {
		resp, err := c.QueryWorkflow(context.Background(), activityInput.WorkflowId, "", "getPlayers")
		var getPlayers map[string]resources.Player

		if err != nil {
			return isPlayer, err
		}

		if err := resp.Get(&getPlayers); err != nil {
			return isPlayer, err
		}

		if _, ok := getPlayers[activityInput.Player]; ok {
			isPlayer = true
		}
	}

	return isPlayer, nil
}
