package triviagame

import (
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func (gs *GameSignal) runPlayerLogic(ctx workflow.Context, workflowInput resources.GameWorkflowInput, getPlayers map[string]resources.Player) bool {
	logger := workflow.GetLogger(ctx)

	// Async timer to cancel game if not started
	cancelTimer := workflow.NewTimer(ctx, time.Duration(workflowInput.StartTimeLimit)*time.Second)

	addPlayerSelector := gs.gameSignal(ctx)

	var cancelTimerFired bool
	addPlayerSelector.AddFuture(cancelTimer, func(f workflow.Future) {
		err := f.Get(ctx, nil)
		if err == nil {
			logger.Info("Time limit for starting game has been exceeded " + time.Duration(workflowInput.StartTimeLimit).String() + " seconds")
			cancelTimerFired = true
		}
	})

	for {
		addPlayerSelector.Select(ctx)

		if gs.Action == "Player" && gs.Player != "" {
			if _, ok := getPlayers[gs.Player]; !ok {
				getPlayers[gs.Player] = resources.Player{
					Score: 0,
				}
			}
		}

		// Wait for start of game via signal
		if gs.Action == "StartGame" {
			break
		}

		if cancelTimerFired {
			return true
		}
	}

	return false
}
