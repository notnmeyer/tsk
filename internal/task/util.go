package task

import "sort"

func alphabetizeTaskList(t *map[string]Task) *[]string {
	// alphabetize task list
	var taskNames []string
	for taskName := range *t {
		taskNames = append(taskNames, taskName)
	}
	sort.Strings(taskNames)
	return &taskNames
}
