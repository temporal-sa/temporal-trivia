package triviagame

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"github.com/sashabaranov/go-openai"
	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func TriviaQuestionActivity(ctx context.Context, input resources.ActivityInput) ([]string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("TriviaQuestionActivity")

	client := openai.NewClient(input.Key)
	messages := make([]openai.ChatCompletionMessage, 0)

	var questions []string
	for q := 0; q < input.NumberOfQuestions; q++ {
		text := "Give me a " + input.Category + " trivia question that has 4 possible answers A), B), C), or D)? Please provide a newline after the question. Give the correct answer such as Answer: A), B), C) or D)"
		text = strings.Replace(text, "\n", "", -1)

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			logger.Warn("ChatCompletion error: %v\n", err)
			q--
			continue
		}

		content := resp.Choices[0].Message.Content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: content,
		})

		logger.Info(content)
		questions = append(questions, content)
		time.Sleep(time.Duration(1) * time.Second)
	}

	return questions, nil
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
