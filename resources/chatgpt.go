package resources

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func SendChatGptRequest(client *openai.Client, messages []openai.ChatCompletionMessage, text string) (string, error) {
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
		fmt.Println(err)
		return "", err
	}

	content := resp.Choices[0].Message.Content
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	})

	return content, nil
}
