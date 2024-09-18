package openai

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// TODO: embed the exmaples file via go:embed
const tskRef = `
	# the environment can be specified at the top level where it is inherited by all tasks
	env = {
	  NAME = "tsk",
	}

	dotenv = ".top.env"

	# the location to look for scripts when a task doesn't contains "cmds"
	# script_dir = "tsk"

	# at its simplest, tasks are a series of sequential shell commands expressed
	# as a list of strings
	[tasks.hello_world]
	cmds = [
	  "echo hello world",
	]

	[tasks.pwd]
	dir = "/tmp" # set the working directory for the task 
	cmds = [
	  "echo \"my pwd is $(pwd)\"", # you can use subshells
	]

	# when "cmds" are omitted tsk attempts to run a script located at "tsk/<task_name>"
	[tasks.no_cmd]
	env = {
	  GREETING = "Hello",
	}

	# tasks can have dependencies. dependencies run before cmds. dependencies are other
	# tasks and cannot be shell commands (yet)
	[tasks.deps]
	deps = [["setup1"]]
	cmds = ["echo 'running cmd...'"]

	# if a task's dep has deps those are run too
	[tasks.deps_of_deps]
	deps = [["setup4"]]
	cmds = ["echo 'running cmd...'"]

	# dependency groups are a way to order dependencies while allowing for parallelization
	[tasks.dep_groups]
	deps = [
	  ["setup1", "setup2"], # setup1 and setup2 run in parallel
	  ["setup3"],           # setup3 runs after the tasks in the previous group complete
	]
	cmds = ["echo 'running cmd...'"]

	# a dotenv file can be supplied at the task level. see the README for information
	# about env var hierarchy
	[tasks.dotenv]
	dotenv = ".env"
	env = {
	  FOO = "bar",
	}
	cmds = [
	  "echo $FOO",
	  "echo $BAR",
	]

	[tasks.top_level_env]
	cmds = [
	  'echo "My name is $NAME!"'
	]

	[tasks.top_level_dotenv]
	cmds = [
	  'echo "$BLAH"'
	]

	[tasks.template]
	cmds = [
	  "echo {{.CLI_ARGS}}"
	]

	# if a dep or command fails tsk exits. in this example, "hello world" will _not_ be echoed.
	[tasks.fail_on_error]
	deps = [["exit"]]
	cmds = ["echo hello world"]

	# tasks used to demonstrate features above
	[tasks.setup1]
	cmds = ["sleep 1", "echo 'doing setup1...'"]

	[tasks.setup2]
	cmds = ["echo 'doing setup2...'"]

	[tasks.setup3]
	cmds = ["echo 'doing setup3...'"]

	[tasks.setup4]
	deps = [["setup2"]]
	cmds = ["echo 'doing setup4...'"]

	[tasks.exit]
	cmds = ["echo exiting 1...", "exit 1"]
`

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
