package task

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"

	"github.com/joho/godotenv"
)

type Vals struct {
	CLI_ARGS string
}

func alphabetizeTaskList(t *map[string]Task) *[]string {
	var taskNames []string
	for taskName := range *t {
		taskNames = append(taskNames, taskName)
	}
	sort.Strings(taskNames)
	return &taskNames
}

func ConvertEnvToStringSlice(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}

func readDotEnv(filename string) (map[string]string, error) {
	dotEnv, err := godotenv.Read(filename)
	if err != nil {
		return nil, err
	}

	return dotEnv, nil
}

func appendDotEnvToEnv(env []string, dotenv string) ([]string, error) {
	additionalEnv, err := readDotEnv(dotenv)
	if err != nil {
		return nil, err
	}
	env = append(env, ConvertEnvToStringSlice(additionalEnv)...)
	return env, nil
}

func render(file, cliArgs string, cliArgsPlaceholder bool) (*bytes.Buffer, error) {
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		return nil, err
	}

	// insert a placeholder value for cliArgs for display purposes
	if cliArgsPlaceholder && cliArgs == "" {
		cliArgs = "{{.CLI_ARGS}}"
	}

	var renderedBuffer bytes.Buffer
	if err := tmpl.Execute(&renderedBuffer, &Vals{CLI_ARGS: cliArgs}); err != nil {
		return nil, err
	}

	return &renderedBuffer, nil
}
