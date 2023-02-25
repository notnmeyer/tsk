package task

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestRunCmd(t *testing.T) {
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	exec.runCommand("echo hello $WORLD", ".", []string{"WORLD=world"})

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

func TestRunTasks(t *testing.T) {
	config := Config{
		Tasks: map[string]Task{
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

	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	err := exec.RunTasks(&config, &[]string{"bar"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	// test the deps run
	foo := regexp.MustCompile(`foo`)
	if !foo.Match(out.Bytes()) {
		t.Errorf("Expected output to contain 'foo', got %s", out.String())
	}
}

func TestDepsRunInParallel(t *testing.T) {
	config := Config{
		Tasks: map[string]Task{
			"one": {
				Cmds: []string{"sleep 1", "echo one"},
			},
			"two": {
				Cmds: []string{"echo two"},
			},
			"zero": {
				Cmds: []string{"echo zero"},
				Deps: [][]string{
					{"one", "two"},
				},
			},
		},
	}

	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	err := exec.RunTasks(&config, &[]string{"zero"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	// test the deps run
	re := regexp.MustCompile(`two\none\nzero`)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected tasks to complete in a specific order (two, one, zero)', got %s", out.String())
	}
}

func TestDepGroupsRunInTheExpectedOrder(t *testing.T) {
	config := Config{
		Tasks: map[string]Task{
			"one": {
				Cmds: []string{"sleep 1", "echo one"},
			},
			"two": {
				Cmds: []string{"echo two"},
			},
			"three": {
				Cmds: []string{"echo three"},
			},
			"zero": {
				Cmds: []string{"echo zero"},
				Deps: [][]string{
					{"one", "two"},
					{"three"},
				},
			},
		},
	}

	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	err := exec.RunTasks(&config, &[]string{"zero"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	// test the deps run
	re := regexp.MustCompile(`two\none\nthree\nzero`)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected tasks to complete in a specific order (two, one, three, zero)', got %s", out.String())
	}
}

func TestFindTaskFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	// testDir should be the project root
	testDir := filepath.Join(cwd, "..", "..", "test", "child")

	err = os.Chdir(testDir)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	path, err := findTaskFile(testDir, "tasks.toml")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if ok, _ := regexp.Match(`tsk/test/tasks.toml`, []byte(path)); !ok {
		t.Errorf("Expected tasks.toml path to match to 'tsk/test/tasks.toml' in %s", path)
	}
}

func TestDotEnv(t *testing.T) {
	cwd, _ := os.Getwd()
	dir := filepath.Join(cwd, "..", "..", "examples")
	os.Chdir(dir)

	var taskFile string
	config, _ := NewTaskConfig(taskFile)

	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	err := exec.RunTasks(config, &[]string{"dotenv"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	expected := "bar\nbaz\n"
	re := regexp.MustCompile(expected)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected %s', got %s", expected, out.String())
	}
}
