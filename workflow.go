package triviagame

import (
	"time"

	"github.com/ktenzer/triviagame/resources"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Trivia game workflow definition
func Workflow(ctx workflow.Context, input resources.Input) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Trivia Game Started")

	// Loop through the number of questions
	for i := 0; i < input.NumberOfQuestions; i++ {

		var result string
		err := workflow.ExecuteActivity(ctx, TriviaQuestionActivity, input).Get(ctx, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}

		logger.Info("Trivia question", "result", result)
	}

	return nil
}
