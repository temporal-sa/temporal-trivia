package triviagame

import (
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func addPlayers(ctx workflow.Context, gameConfiguration *resources.GameConfiguration, getPlayers *map[string]resources.Player) bool {
	logger := workflow.GetLogger(ctx)

	// Async timer to cancel game if not started
	timerCtx, cancelTimer := workflow.WithCancel(ctx)
	startGameTimer := workflow.NewTimer(timerCtx, time.Duration(gameConfiguration.StartTimeLimit)*time.Second)
	startGameTimerSelector := workflow.NewSelector(ctx)

	var cancelTimerFired bool
	startGameTimerSelector.AddFuture(startGameTimer, func(f workflow.Future) {
		err := f.Get(timerCtx, nil)
		if err == nil {
			logger.Info("Time limit for starting game has been exceeded " + intToString(gameConfiguration.StartTimeLimit) + " seconds")
			cancelTimerFired = true
		}
	})

	isStartGame := boolPointer(false)

	// Receive signal using workflow coroutine async (for fun)
	workflow.Go(ctx, func(gCtx workflow.Context) {
		gs := GameSignal{}

		startGameSelector := workflow.NewSelector(gCtx)
		gs.gameSignal(ctx, startGameSelector)

		startGameSelector.Select(gCtx)

		if gs.Action == "StartGame" {
			*isStartGame = true
		}
	})

	// Wait for either timer fired or start game signal
	workflow.Await(ctx, func() bool {
		if *isStartGame {
			return true
		}

		if cancelTimerFired {
			return true
		}

		return false
	})

	// return back to workflow
	if cancelTimerFired {
		return true
	} else {
		// cancel timer
		cancelTimer()

		return false
	}
}

func boolPointer(b bool) *bool {
	return &b
}
