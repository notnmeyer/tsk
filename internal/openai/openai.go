package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	_ "embed"

	openai "github.com/sashabaranov/go-openai"
)

type prompt struct {
	Base            string `json:"base"`             // the base prompt
	TaskName        string `json:"task_name"`        // the name of the generated task
	TaskDesc        string `json:"task_desc"`        // the description of what the task does
	FormatReference string `json:"format_reference"` // the reference of tsk's toml format
}

//go:embed _reference_tasks.toml
var formatReference string

//go:embed prompt/base.md
var promptBase string

var p = &prompt{
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

func GenerateTask(name, taskDesc string) (*string, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	p.TaskName = name
	p.TaskDesc = taskDesc

	content, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: string(content),
			},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("completion error: %v\n", err)
	}

	return &resp.Choices[0].Message.Content, nil
}
