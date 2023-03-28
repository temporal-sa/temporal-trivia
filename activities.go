package triviagame

import (
	"context"
	"strconv"

	"github.com/ktenzer/temporal-trivia/resources"
	"github.com/sashabaranov/go-openai"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func TriviaQuestionActivity(ctx context.Context, input resources.ActivityInput) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("TriviaQuestionActivity")

	client := openai.NewClient(input.Key)
	messages := make([]openai.ChatCompletionMessage, 0)

	result, err := resources.SendChatGptRequest(client, messages, input.Question)

	if err != nil {
		return result, temporal.NewApplicationError("ChatGPT", "request", err)
	}

	return result, nil
}

func ScoreTotalActivity(ctx context.Context, scoreboardMap map[string]int) ([]string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("ScoreTotalActivity")

	var highestScore int
	var playersWithHighestScore []string

	for player, score := range scoreboardMap {
		if score > highestScore {
			highestScore = score
			playersWithHighestScore = []string{player + ":" + strconv.Itoa(highestScore)}
		} else if score == highestScore {
			playersWithHighestScore = append(playersWithHighestScore, player+":"+strconv.Itoa(highestScore))
		}
	}

	return playersWithHighestScore, nil
}
