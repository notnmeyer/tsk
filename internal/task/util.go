package task

import (
	"fmt"
	"github.com/joho/godotenv"
	// "path/filepath"
	"sort"
	"strings"
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

func indent(count int) string {
	return strings.Repeat(" ", count)
}
