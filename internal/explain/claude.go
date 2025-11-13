package explain

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// callClaude sends a prompt to Claude API and returns the response
func callClaude(ctx context.Context, prompt string, apiKey string) (string, error) {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
		MaxTokens: anthropic.F(int64(4096)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return "", fmt.Errorf("claude api error: %w", err)
	}

	// Extract the text from the response
	if len(message.Content) > 0 {
		content := message.Content[0]
		return content.Text, nil
	}

	return "", fmt.Errorf("no response from claude")
}
