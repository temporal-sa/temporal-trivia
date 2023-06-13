package triviagame

import (
	"context"
	"sort"

	"github.com/ktenzer/temporal-trivia/resources"
	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func LeaderBoardActivity(ctx context.Context, getPlayers map[string]resources.Player) ([]resources.ScoreBoard, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ScoreTotalActivity")

	scoreboard := sortLeaderBoard(getPlayers)

	return scoreboard, nil
}

func sortLeaderBoard(getPlayers map[string]resources.Player) []resources.ScoreBoard {
	var scoreboard []resources.ScoreBoard
	for k, _ := range getPlayers {
		scoreboard = append(scoreboard, resources.ScoreBoard{Player: k, Score: getPlayers[k].Score})
	}

	sort.Slice(scoreboard, func(i, j int) bool {
		return scoreboard[i].Score > scoreboard[j].Score
	})

	return scoreboard
}
