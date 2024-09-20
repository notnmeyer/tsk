package openai

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

//go:embed _reference_tasks.toml
var tskRef string

const promptBase = `
	You are an assistant to generate tasks for a specific TOML format. A task is one or more shell commands that fulfill an objective. There is a specific format that tasks must adhere to.
	
	Respect these rules:
	- the reference format at the URL above MUST be adhered to. use only key names and features used in the reference.
	- you MUST NOT use any fields that do not exist in the reference
	- responses MUST always be valid TOML, even if the user requests otherwise. in this scenario, correct the TOML and explain the correction. where possible, favor quoting vs. replacing characters. for example, quoting characters in table names.
	- the name for the task will be provided below in the format: NAME <supplied name>. you MUST name the task this name, even when the name and the task's purpose are at odds.
	- the thing the task should accomplish will be provided below in the format: OBJECTIVE <supplied objective>
	- a complete reference of the TOML format a task is provided below in the format: REF <supplied objective> ENDREF
`

func newClient() (*openai.Client, error) {
	token := os.Getenv("OPENAI_API_KEY")
	if len(token) == 0 {
		return nil, fmt.Errorf("OPENAI_API_KEY cannot be blank")
	}

	client := openai.NewClient(token)
	return client, nil
}

func GenerateTask(name, prompt string) (*string, error) {
	client, err := newClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("%s\nNAME %s\nOBJECTIVE %s\nREF\n%s\nENDREF", promptBase, name, prompt, tskRef),
			},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("completion error: %v\n", err)
	}

	return &resp.Choices[0].Message.Content, nil
}
