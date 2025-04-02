package main

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name            string
		input           []string
		expectedTasks   []string
		expectedCliArgs string
		dashIndex       int
	}{
		{
			name:            "tasks and args",
			input:           []string{"task1", "task2", "arg1", "arg2"},
			expectedTasks:   []string{"task1", "task2"},
			expectedCliArgs: "arg1 arg2",
			dashIndex:       2,
		},
		{
			name:            "just tasks",
			input:           []string{"task1", "task2"},
			expectedTasks:   []string{"task1", "task2"},
			expectedCliArgs: "",
			dashIndex:       -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tasks, cliArgs := parseArgs(test.input, test.dashIndex)
			if !equalSlices(tasks, test.expectedTasks) {
				t.Errorf("Expected tasks: %v, got: %v", test.expectedTasks, tasks)
			}
			if cliArgs != test.expectedCliArgs {
				t.Errorf("Expected cliArgs: %q, got: %q", test.expectedCliArgs, cliArgs)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
