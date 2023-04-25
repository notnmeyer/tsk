package task

import (
	"fmt"
	"sort"
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
