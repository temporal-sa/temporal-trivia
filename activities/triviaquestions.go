package triviagame

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"github.com/sashabaranov/go-openai"
	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func TriviaQuestionActivity(ctx context.Context, input resources.ActivityInput) (map[int]resources.Result, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("TriviaQuestionActivity")

	// log message we encountered heartbeat timeout
	var completedQuestion int
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &completedQuestion); err == nil {
			logger.Info("Resuming from failed attempt", "ReportedProgress", completedQuestion)
		}
	}

	// openai client
	client := openai.NewClient(input.Key)
	messages := make([]openai.ChatCompletionMessage, 0)

	// pre-fetch list of questions
	var questions []string
	for q := 0; q < input.NumberOfQuestions; q++ {
		activity.RecordHeartbeat(ctx, q)

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

	gameMap := populateGameMap(questions)
	return gameMap, nil
}

// populate gameMap
func populateGameMap(questions []string) map[int]resources.Result {
	gameMap := make(map[int]resources.Result)
	for i, question := range questions {
		var result resources.Result
		correctAnswer := parseCorrectAnswer(question)
		result.Answer = correctAnswer

		result.Question = parseQuestion(question)
		answersMap := parsePossibleAnswers(question)
		result.MultipleChoiceMap = answersMap

		gameMap[i] = result
	}

	return gameMap
}

// Parse the question
func parseQuestion(question string) string {
	questionRegex := regexp.MustCompile(`^(.*?\?)\s*\nA\)|B\)|C\)|D\)`)
	match := questionRegex.FindStringSubmatch(question)

	var parsedQuestion string
	if len(match) > 1 {
		parsedQuestion = match[1]
	}

	return parsedQuestion
}

// Parse the possible answers
func parsePossibleAnswers(question string) map[string]string {
	re := regexp.MustCompile(`([A-Z])\) (\w+(?: \w+)*)`)
	matches := re.FindAllStringSubmatch(question, -1)

	answers := make(map[string]string)
	for _, match := range matches {
		answers[match[1]] = match[2]
	}

	return answers
}

// Parse answer
func parseCorrectAnswer(question string) string {
	re := regexp.MustCompile(`\w+\s*Answer:? ([A-D])\)?`)

	match := re.FindStringSubmatch(question)
	if len(match) > 0 {
		return match[1]
	}

	return ""
}
