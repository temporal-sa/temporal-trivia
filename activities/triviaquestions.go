package triviagame

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ktenzer/temporal-trivia/resources"
	"github.com/sashabaranov/go-openai"
	"go.temporal.io/sdk/activity"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

func TriviaQuestionKapaAI(ctx context.Context, input resources.TriviaQuestionsActivityInput) (map[int]resources.Result, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("KAPA AI TriviaQuestionActivity")

	var gameMap map[int]resources.Result
	var threadId string
	var questions []string
	for q := 0; q < input.NumberOfQuestions; q++ {
		category, err := GetRandomTemporalCategoryActivity(ctx)
		if err != nil {
			return gameMap, err
		}

		var text string
		var apiURL string
		text = "Please+provide+me+with+a+new+random+" + category + "+Temporal+question.+followed+by+four+possible+answers+in+the+format+A),+B),+C),+D).+Please+start+a+new+line+after+the+question.+Also,+provide+the+correct+answer+in+the+format+'Answer:+A)',+'Answer:+B)',+'Answer:+C)'+or+'Answer:+D)'.+Please+ensure+that+there+is+a+space+between+the+colon+and+the+answer+letter+in+your+response."
		text = strings.Replace(text, "\n", "", -1)

		if len(strings.TrimSpace(threadId)) == 0 {
			apiURL = "https://api.kapa.ai/query/v1?query=" + text
		} else {
			apiURL = "https://api.kapa.ai/query/v1/thread/f9b453b4-dd78-414e-9177-95edbd48a22c?query=" + text
		}

		apiKey := os.Getenv("KAPA_API_KEY")

		// Create a new request using http
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			logger.Error("Error creating request:", err)
			return gameMap, err
		}

		// Add API Token
		req.Header.Set("X-API-TOKEN", apiKey)

		// Create a client and send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error making request:", err)
			return gameMap, err
		}
		defer resp.Body.Close()

		// Check the response status code
		if resp.StatusCode != http.StatusOK {
			logger.Error("Received non-200 response code: %d, status: %s\n", resp.StatusCode, resp.Status)
			return gameMap, err
		}

		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Error reading response body:", err)
			return gameMap, err
		}

		// Decode the JSON response
		var kapaResp resources.KapaResponse
		err = json.Unmarshal(bodyBytes, &kapaResp)
		if err != nil {
			logger.Error("Error decoding JSON response:", err)
			return gameMap, err
		}

		logger.Info(kapaResp.Answer)
		questions = append(questions, kapaResp.Answer)
		threadId = kapaResp.ThreadID
	}

	gameMap = populateGameMap(questions)
	return gameMap, nil
}

func TriviaQuestionChatGPT(ctx context.Context, input resources.TriviaQuestionsActivityInput) (map[int]resources.Result, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("ChatGPT TriviaQuestionActivity")

	// openai client
	client := openai.NewClient(os.Getenv("CHATGPT_API_KEY"))
	messages := make([]openai.ChatCompletionMessage, 0)

	// pre-fetch list of questions
	var questions []string
	for q := 0; q < input.NumberOfQuestions; q++ {
		//activity.RecordHeartbeat(ctx, q)

		text := "Please provide me with a " + input.Category + " trivia question, followed by four possible answers in the format A), B), C), D). Please start a new line after the question. Also, provide the correct answer in the format 'Answer: A)', 'Answer: B)', 'Answer: C)' or 'Answer: D)'. Please ensure that there is a space between the colon and the answer letter in your response."
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

		gameMap[i+1] = result
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
	re := regexp.MustCompile(`([A-Z])\) (\S+(?: \S+)*)`)
	matches := re.FindAllStringSubmatch(question, -1)

	answers := make(map[string]string)
	for _, match := range matches {
		answers[match[1]] = match[2]
	}

	return answers
}

// Parse answer
func parseCorrectAnswer(question string) string {
	re := regexp.MustCompile(`\s*Answer:? ([A-D])\)?`)

	match := re.FindStringSubmatch(question)
	if len(match) > 0 {
		return match[1]
	}

	return ""
}
