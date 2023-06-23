package triviagame

import (
	"sort"
	"strconv"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Detect answer duplication
func isAnswerDuplicate(submissionMap map[string]resources.Submission, player string) bool {
	if _, ok := submissionMap[player]; ok {
		return true
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

// Triva Questions Activity Input
func triviaQuestionsActivityInput(category string, numberQuestions int) resources.TriviaQuestionsActivityInput {
	activityInput := resources.TriviaQuestionsActivityInput{
		Category:          category,
		NumberOfQuestions: numberQuestions,
	}

	return activityInput
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

func intToString(i int) string {
	return strconv.Itoa(i)
}
