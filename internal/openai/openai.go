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
	BaseInstruction string `json:"base"`      // the base prompt
	TaskName        string `json:"task_name"` // the name of the generated task
	TaskDesc        string `json:"task_desc"` // the description of what the task does
	TaskRef         string `json:"task_ref"`  // the reference of tsk's toml format
}

//go:embed _reference_tasks.toml
var tskRef string

var p = &prompt{
	BaseInstruction: `
		You are an assistant generating tasks for the program "tsk" in TOML. A "task" is typically one or more shell commands that do something like run tests or build and deploy an app. There is a specific format that tasks you generate must adhere to.
	
		Respect these rules:
		- responses MUST always be valid TOML, even if the user requests otherwise. correct the TOML if necessary. where possible, favor quoting vs. replacing characters. for example, quoting characters in table names.
		- the name for the task will be provided in the "task_name" key. you MUST name the task this name, even when the name and the task's purpose are at odds.
		- the description of what the task should do is provided in the "task_desc" key.
		- the reference of all available features for a task is provided in a set of example tasks provided in the "task_ref" key.
			- use only key names and features used in the reference.
			- you MUST NOT use any fields that do not exist in the reference.
		- favor using sh or bash shell features. exceptions are fine where using an external program is required or if it makes a task significantly shorter, simpler, or easier to express.
		- response ONLY with the TOML for the task. do not provide any additional explanation or commentary.
		- DO NOT wrap the task's TOML in yaml, or wrap responses in YAML code blocks. respond only with the TOML.
	`,
	TaskRef: tskRef,
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
