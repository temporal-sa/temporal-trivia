package triviagame

import (
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func (ps *PlayerSignal) runPlayerLogic(ctx workflow.Context, workflowInput resources.GameWorkflowInput, getPlayers *map[string]resources.Player) bool {
	logger := workflow.GetLogger(ctx)

	// Async timer to cancel game if not started
	cancelTimer := workflow.NewTimer(ctx, time.Duration(workflowInput.StartTimeLimit)*time.Second)
	addPlayerSelector := workflow.NewSelector(ctx)
	ps.playerSignal(ctx, addPlayerSelector)

	//ps := &signal.playerSignal

	var cancelTimerFired bool
	addPlayerSelector.AddFuture(cancelTimer, func(f workflow.Future) {
		err := f.Get(ctx, nil)
		if err == nil {
			logger.Info("Time limit for starting game has been exceeded " + time.Duration(workflowInput.StartTimeLimit).String() + " seconds")
			cancelTimerFired = true
		}
	})

	isStartGame := boolPointer(false)
	workflow.Go(ctx, func(gCtx workflow.Context) {
		gs := GameSignal{}

		startGameSelector := workflow.NewSelector(gCtx)
		gs.gameSignal(ctx, startGameSelector)

		startGameSelector.Select(gCtx)

		if gs.Action == "StartGame" {
			*isStartGame = true
		}
	})

	workflow.Go(ctx, func(gCtx workflow.Context) {
		for {
			addPlayerSelector.Select(gCtx)

			if ps.Action == "Player" && ps.Player != "" {
				if _, ok := (*getPlayers)[ps.Player]; !ok {
					(*getPlayers)[ps.Player] = resources.Player{
						Score: 0,
					}
				}
			}
		}
	})

	workflow.Await(ctx, func() bool {
		if *isStartGame {
			return true
		}

		if cancelTimerFired {
			return true
		}

		return false
	})

	return false
}

func boolPointer(b bool) *bool {
	return &b
}
