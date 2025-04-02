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

func TestConfig_CompileEnv(t *testing.T) {
	t.Run("without dotenv", func(t *testing.T) {
		config := Config{
			Env: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
			DotEnv:      "",
			TaskFileDir: "/some/path",
		}

		env, err := config.CompileEnv()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		expectedEnv := []string{
			"FOO=bar",
			"BAZ=qux",
		}

		if len(env) != len(expectedEnv) {
			t.Fatalf("expected %d env vars, got %d", len(expectedEnv), len(env))
		}

		for _, e := range expectedEnv {
			found := false
			for _, actual := range env {
				if e == actual {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected env var %s not found", e)
			}
		}
	})

	t.Run("with dotenv", func(t *testing.T) {
		dotEnvContent := "FOO=bar_from_dotenv\nBAZ=qux_from_dotenv\n"
		dotEnvPath := createTempDotEnv(t, dotEnvContent)
		defer removeFile(t, dotEnvPath)

		config := Config{
			Env: map[string]string{
				"FOO": "bar",
			},
			DotEnv:      filepath.Base(dotEnvPath),
			TaskFileDir: filepath.Dir(dotEnvPath),
		}

		env, err := config.CompileEnv()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		expectedEnv := []string{
			"FOO=bar_from_dotenv", // overwritten by dotenv
			"BAZ=qux_from_dotenv",
		}

		for _, e := range expectedEnv {
			found := false
			for _, actual := range env {
				if e == actual {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected env var %s not found", e)
			}
		}
	})

	t.Run("error loading dotenv", func(t *testing.T) {
		config := Config{
			DotEnv:      "nonexistent.env",
			TaskFileDir: "/some/nonexistent/path",
		}

		_, err := config.CompileEnv()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestTask_CompileEnv(t *testing.T) {
	t.Run("with task-specific env and inherited environment", func(t *testing.T) {
		baseEnv := []string{"GLOBAL=global_value"}
		task := Task{
			Env: map[string]string{
				"TASK_KEY": "task_value",
			},
			Pure: false,
		}

		compiledEnv, err := task.CompileEnv(baseEnv)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		expectedEnv := []string{
			"GLOBAL=global_value",
			"TASK_KEY=task_value",
		}

		for _, expected := range expectedEnv {
			found := false
			for _, actual := range compiledEnv {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected env var %s not found", expected)
			}
		}
	})

	t.Run("pure environment", func(t *testing.T) {
		task := Task{
			Env: map[string]string{
				"TASK_KEY": "task_value",
			},
			Pure: true,
		}

		compiledEnv, err := task.CompileEnv([]string{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		expectedEnv := []string{
			"TASK_KEY=task_value",
			"USER=" + os.Getenv("USER"),
			"HOME=" + os.Getenv("HOME"),
		}

		for _, expected := range expectedEnv {
			found := false
			for _, actual := range compiledEnv {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected env var %s not found", expected)
			}
		}
	})

	t.Run("with dotenv", func(t *testing.T) {
		dotEnvContent := "DOTENV_KEY=dotenv_value\n"
		dotEnvPath := createTempDotEnv(t, dotEnvContent)
		defer removeFile(t, dotEnvPath)

		task := Task{
			DotEnv: filepath.Base(dotEnvPath),
			Dir:    filepath.Dir(dotEnvPath),
		}

		compiledEnv, err := task.CompileEnv([]string{})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		expectedEnv := []string{
			"DOTENV_KEY=dotenv_value",
		}

		for _, expected := range expectedEnv {
			found := false
			for _, actual := range compiledEnv {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected env var %s not found in result", expected)
			}
		}
	})

	t.Run("error loading dotenv", func(t *testing.T) {
		task := Task{
			DotEnv: "nonexistent.env",
			Dir:    "/some/nonexistent/path",
		}

		_, err := task.CompileEnv([]string{})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
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
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: &Config{
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
		},
	}

	err := exec.RunTasks(exec.Config, &[]string{"bar"})
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
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: &Config{
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
		},
	}

	err := exec.RunTasks(exec.Config, &[]string{"zero"})
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
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: &Config{
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
		},
	}

	err := exec.RunTasks(exec.Config, &[]string{"zero"})
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
	var taskFile, cliArgs string

	config, err := NewTaskConfig(taskFile, cliArgs, false)
	if err != nil {
		panic(err)
	}
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: config,
	}

	err = exec.RunTasks(exec.Config, &[]string{"dotenv"})
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
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: &Config{
			Tasks: map[string]Task{
				"default": {
					// examples/.env sets BAR=baz
					DotEnv: ".env",
					Env:    map[string]string{"BAR": "baz2"},
					Cmds:   []string{"echo $BAR"},
				},
			},
		},
	}

	err := exec.RunTasks(exec.Config, &[]string{"default"})
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
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: &Config{
			Env: map[string]string{"BAR": expected},
			Tasks: map[string]Task{
				"default": {
					Cmds: []string{"echo $BAR"},
				},
			},
		},
	}

	exec.RunTasks(exec.Config, &[]string{"default"})
	re := regexp.MustCompile(expected)
	if !re.Match(out.Bytes()) {
		t.Errorf("Expected '%s', got %s", expected, out.String())
	}
}

// a task's env should override the global env
func TestGlobalEnvInheritance(t *testing.T) {
	expected := "baz2"
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: &Config{
			Env: map[string]string{"BAR": "baz"},
			Tasks: map[string]Task{
				"default": {
					Env:  map[string]string{"BAR": expected},
					Cmds: []string{"echo $BAR"},
				},
			},
		},
	}

	exec.RunTasks(exec.Config, &[]string{"default"})
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
	config, _ := NewTaskConfig(path, cliArgs, false)
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: config,
	}

	exec.RunTasks(exec.Config, &[]string{"template"})
	if !expected.Match(out.Bytes()) {
		t.Errorf("Expected '%s' to match '%s'", cliArgs, out.String())
	}
}

func TestRunTasksWithInvalidDependency(t *testing.T) {
	exec := Executor{
		Stdout: new(bytes.Buffer),
		Config: &Config{
			Tasks: map[string]Task{
				"foo": {
					Deps: [][]string{{"non-existent-task"}},
					Cmds: []string{"echo foo"},
				},
			},
		},
	}

	err := exec.RunTasks(exec.Config, &[]string{"foo"})
	if err == nil {
		t.Error("Expected error for non-existent dependency, got nil")
	}
}

// when building --list output for tasks that use CLI_ARGS test that placeholder
// text is inserted when CLI_ARGS arent provided
func TestTemplatesWithPlaceholders(t *testing.T) {
	placeholder := "{{.CLI_ARGS}}"
	expected := regexp.MustCompile(placeholder)
	wd, _ := os.Getwd()
	path, _ := findTaskFile(wd, "tasks.toml")
	config, _ := NewTaskConfig(path, "", true)
	out := new(bytes.Buffer)
	exec := Executor{
		Stdout: out,
		Config: config,
	}

	exec.RunTasks(exec.Config, &[]string{"template"})
	if !expected.Match(out.Bytes()) {
		t.Errorf("Expected '%s' to match '%s'", placeholder, out.String())
	}
}

//
// helpers
//

// helper for creating .env
func createTempDotEnv(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test_dotenv_*.env")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()
	return tmpFile.Name()
}

// helper to remove a file
func removeFile(t *testing.T, path string) {
	t.Helper()
	err := os.Remove(path)
	if err != nil {
		t.Fatalf("failed to rm file %s: %v", path, err)
	}
}
