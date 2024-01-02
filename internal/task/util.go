package task

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

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

func render(file string) (*bytes.Buffer, error) {
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		return nil, err
	}

	var renderedBuffer bytes.Buffer

	// TODO: how and where template values come from
	vals := make(map[string]string)

	var cliArgs []string
	if FindIndex(os.Args) > 0 {
		cliArgs = os.Args[FindIndex(os.Args)+1:]
	}

	vals["CLI_ARGS"] = strings.Join(cliArgs, " ")

	if err := tmpl.Execute(&renderedBuffer, vals); err != nil {
		return nil, err
	}

	return &renderedBuffer, nil
}

func FindIndex(args []string) int {
	for i, arg := range args {
		if arg == "--" {
			return i
		}
	}
	return -1
}
