package triviagame

import (
	"context"
	"math/rand"

	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func GetRandomCategoryActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GetRandomCategoryActivity")

	keys := []string{"Temporal.io", "General", "Sports", "Science", "Travel", "Geography", "Capitols", "Authors", "Books", "Animals", "Plants", "Foods", "Cities"}
	randomIndex := rand.Intn(len(keys))

	logger.Info("Category is " + keys[randomIndex])

	return keys[randomIndex], nil
}

func GetRandomTemporalCategoryActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("GetRandomCategoryActivity")

	keys := []string{"SDK", "Cloud", "CLI", "History", "Server", "Worker", "General"}
	randomIndex := rand.Intn(len(keys))

	logger.Info("Temporal Category is " + keys[randomIndex])

	return keys[randomIndex], nil
}
