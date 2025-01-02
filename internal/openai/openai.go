package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	_ "embed"

	openai "github.com/sashabaranov/go-openai"
)

type systemPrompt struct {
	Base            string `json:"base"`             // the base prompt
	FormatReference string `json:"format_reference"` // the reference of tsk's toml format
}

type userPrompt struct {
	TaskName string `json:"task_name"` // the name of the generated task
	TaskDesc string `json:"task_desc"` // the description of what the task does
}

//go:embed _reference_tasks.toml
var formatReference string

//go:embed prompt/base.md
var promptBase string

var sp = &systemPrompt{
	Base:            promptBase,
	FormatReference: formatReference,
}

func newClient() (*openai.Client, error) {
	token := os.Getenv("OPENAI_API_KEY")
	if len(token) == 0 {
		return nil, fmt.Errorf("to generate a task, an OpenAI API key must be set to the env var OPENAI_API_KEY")
	}

	client := openai.NewClient(token)
	return client, nil
}

func GenerateTask(name, desc string) (*string, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	// prepare the system prompt
	system, err := json.Marshal(sp)
	if err != nil {
		return nil, err
	}

	// prepare the user prompt
	user, err := json.Marshal(&userPrompt{
		TaskName: name,
		TaskDesc: desc,
	})

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4oLatest,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: string(system),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: string(user),
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("completion error: %v\n", err)
	}

	return &resp.Choices[0].Message.Content, nil
}
