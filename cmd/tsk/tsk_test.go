package main

import (
	"testing"

	"github.com/notnmeyer/tsk/internal/task"
)

func TestVerifyTasks(t *testing.T) {
	config := task.Config{
		Tasks: map[string]task.Task{
			"foo": {
				Cmds: []string{"echo foo"},
			},
			"bar": {
				Cmds: []string{"echo bar"},
				Deps: [][]string{
					{"foo"},
				},
			},
		},
	}

	err := verifyTasks(&config, []string{"foo", "bar"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	err = verifyTasks(&config, []string{"foo", "baz"})
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
