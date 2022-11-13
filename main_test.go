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
	actual := ConvertEnv(env)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(actual))
	}

	for i, v := range actual {
		if v != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], v)
		}
	}
}
