package triviagame

import (
	"context"
	"sort"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func LeaderBoardActivity(ctx context.Context, scoreMap map[string]int) ([]resources.ScoreBoard, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ScoreTotalActivity")

	scoreboard := sortLeaderBoard(scoreMap)

	return scoreboard, nil
}

func sortLeaderBoard(scoreMap map[string]int) []resources.ScoreBoard {
	var scoreboard []resources.ScoreBoard
	for k, v := range scoreMap {
		scoreboard = append(scoreboard, resources.ScoreBoard{Player: k, Score: v})
	}

	sort.Slice(scoreboard, func(i, j int) bool {
		return scoreboard[i].Score > scoreboard[j].Score
	})

	return scoreboard
}
