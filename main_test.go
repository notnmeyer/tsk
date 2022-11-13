package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRunCmd(t *testing.T) {
	out := new(bytes.Buffer)
	opts := Opts{
		Stdout: out,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}

	RunCommand("echo hello $WORLD", ".", []string{"WORLD=world"}, opts)

	if out.String() != "hello world\n" {
		t.Errorf("Expected 'hello world', got %s", out)
	}
}

func TestConvertEnv(t *testing.T) {
	env := map[string]string{
		"FOO": "bar",
	}

	expected := []string{"FOO=bar"}
	actual := ConvertEnvToStringSlice(env)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(actual))
	}

	for i, v := range actual {
		if v != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], v)
		}
	}
}

func TestVerifyTasks(t *testing.T) {
	config := taskConfig{
		Tasks: map[string]Task{
			"foo": {
				Cmds: []string{"echo foo"},
			},
			"bar": {
				Cmds: []string{"echo bar"},
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

func TestConvertEnvToStringSlice(t *testing.T) {
	env := map[string]string{
		"FOO": "bar",
	}

	expected := []string{"FOO=bar"}
	actual := ConvertEnvToStringSlice(env)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(actual))
	}

	for i, v := range actual {
		if v != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], v)
		}
	}
}
