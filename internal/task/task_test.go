package task

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestMain(m *testing.M) {
	// some tests will load files from ./examples/. for convenience, change the wd to this directory
	cwd, _ := os.Getwd()
	dir := filepath.Join(cwd, "..", "..", "examples")
	os.Chdir(dir)
	os.Exit(m.Run())
}

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

	// test the deps TestRunTasks
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

// find test/tasks.toml from test/child/
func TestFindTaskFile(t *testing.T) {
	cwd, _ := os.Getwd()
	testDir := filepath.Join(cwd, "..", "test", "child")

	os.Chdir(testDir)
	defer os.Chdir(cwd)

	path, err := findTaskFile(testDir, "tasks.toml")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if ok, _ := regexp.Match(`tsk/test/tasks.toml`, []byte(path)); !ok {
		t.Errorf("Expected tasks.toml path to match to 'tsk/test/tasks.toml' in %s", path)
	}
}

// test .env file is loaded
func TestDotEnv(t *testing.T) {
	var (
		taskFile string
		cliArgs  string
	)
	config, err := NewTaskConfig(taskFile, cliArgs)
	if err != nil {
		panic(err)
	}
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	err = exec.RunTasks(config, &[]string{"dotenv"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	expected := "bar\nbaz\n"
	re := regexp.MustCompile(expected)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected %s', got %s", expected, out.String())
	}
}

// `task_name.DotEnv` overrides `task_name.Env`
func TestEnvInheritance(t *testing.T) {
	expected := "baz"
	config := Config{
		Tasks: map[string]Task{
			"default": {
				// examples/.env sets BAR=baz
				DotEnv: ".env",
				Env:    map[string]string{"BAR": "baz2"},
				Cmds:   []string{"echo $BAR"},
			},
		},
	}

	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	err := exec.RunTasks(&config, &[]string{"default"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	re := regexp.MustCompile(expected)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected '%s', got %s", expected, out.String())
	}
}

func TestGlobalEnv(t *testing.T) {
	expected := "baz2"
	config := Config{
		Env: map[string]string{"BAR": expected},
		Tasks: map[string]Task{
			"default": {
				Cmds: []string{"echo $BAR"},
			},
		},
	}

	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	exec.RunTasks(&config, &[]string{"default"})
	re := regexp.MustCompile(expected)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected '%s', got %s", expected, out.String())
	}
}

// a task's env should override the global env
func TestGlobalEnvInheritance(t *testing.T) {
	expected := "baz2"
	config := Config{
		Env: map[string]string{"BAR": "baz"},
		Tasks: map[string]Task{
			"default": {
				Env:  map[string]string{"BAR": expected},
				Cmds: []string{"echo $BAR"},
			},
		},
	}

	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	exec.RunTasks(&config, &[]string{"default"})
	re := regexp.MustCompile(expected)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected '%s', got %s", expected, out.String())
	}
}

func TestFilterTasks(t *testing.T) {
	var (
		tasks = map[string]Task{
			"foo": {},
			"bar": {},
		}
		expectedKey = "foo"
		re          = regexp.MustCompile(expectedKey)
		result      = filterTasks(&tasks, re)
	)

	if len(result) != 1 {
		t.Errorf("Expected `len(res) == 1`, got %d", len(result))
	}

	if _, ok := result[expectedKey]; !ok {
		t.Errorf("Expected key %s to exist", expectedKey)
	}
}

// CLI_ARGS template
func TestTemplates(t *testing.T) {
	cliArgs := "foobar"
	expected := regexp.MustCompile(cliArgs)
	wd, _ := os.Getwd()
	path, _ := findTaskFile(wd, "tasks.toml")
	config, _ := NewTaskConfig(path, cliArgs)
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
	}

	exec.RunTasks(config, &[]string{"template"})
	if !expected.Match(out.Bytes()) {
		t.Errorf("Expected '%s' to match '%s'", cliArgs, out.String())
	}
}
