package triviagame

import (
	"sort"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Detect answer duplication
func isAnswerDuplicate(submissions map[string]resources.Submission, player string) bool {
	for submittedPlayer, _ := range submissions {
		if player == submittedPlayer {
			return true
		}
	}
	return false
}

// Sort gameMap
func getSortedGameMap(gameMap map[int]resources.Result) []int {

	keys := make([]int, 0, len(gameMap))
	for k := range gameMap {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	return keys
}

// Activity Options
func setDefaultActivityOptions() workflow.ActivityOptions {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 300 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}

	return ao
}

// Local Activity Options
func setDefaultLocalActivityOptions() workflow.LocalActivityOptions {
	ao := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			BackoffCoefficient: 1.0,
			MaximumAttempts:    5,
		},
	}

	return ao
}
