package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// represents parsed task file
type Config struct {
	DotEnv      string
	Env         map[string]string
	Tasks       map[string]Task
	ScriptDir   string
	TaskFileDir string
}

// represents an individual task
type Task struct {
	Cmds   []string
	Deps   [][]string
	Dir    string
	Env    map[string]string
	DotEnv string
	Pure   bool
}

type Executor struct {
	Stdout io.Writer
	Stdin  io.Reader
	Stderr io.Writer
	Config *Config
}

// sets the top-level env
func (c *Config) CompileEnv() ([]string, error) {
	env := ConvertEnvToStringSlice(c.Env)

	if c.DotEnv != "" {
		var err error
		env, err = appendDotEnvToEnv(env, filepath.Join(c.TaskFileDir, c.DotEnv))
		if err != nil {
			return nil, err
		}
	}

	return env, nil
}

func (t *Task) CompileEnv(env []string) ([]string, error) {
	env = append(env, ConvertEnvToStringSlice(t.Env)...)

	// a "pure" environment does not inherit the full parent env, but does inherit
	// USER and HOME. otherwise it inherits the entire parent env.
	if t.Pure {
		env = append(
			env,
			fmt.Sprintf("USER=%s", os.Getenv("USER")),
			fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		)
	} else {
		env = append(env, os.Environ()...)
	}

	if t.DotEnv != "" {
		var err error
		env, err = appendDotEnvToEnv(env, filepath.Join(t.Dir, t.DotEnv))
		if err != nil {
			return nil, err
		}
	}

	return env, nil
}

func (exec *Executor) RunTasks(config *Config, tasks *[]string) error {
	// top-level env
	env, err := config.CompileEnv()
	if err != nil {
		return err
	}

	for _, task := range *tasks {
		taskConfig := config.Tasks[task]

		if taskConfig.Dir == "" {
			taskConfig.Dir = config.TaskFileDir
		}

		if len(taskConfig.Deps) > 0 {
			for _, depGroup := range taskConfig.Deps {
				var wg sync.WaitGroup
				wg.Add(len(depGroup))
				for _, dep := range depGroup {
					go func(dep string) {
						defer wg.Done()
						exec.RunTasks(config, &[]string{dep})
					}(dep)
				}
				wg.Wait()
			}
		}

		// add any task-specific env bits
		env, err = taskConfig.CompileEnv(env)
		if err != nil {
			return err
		}

		// if a task contains cmds, run them
		if len(taskConfig.Cmds) > 0 {
			for _, cmd := range taskConfig.Cmds {
				err := exec.runCommand(cmd, taskConfig.Dir, env)
				if err != nil {
					return err
				}
			}
		} else {
			// if there are no cmds then we intend to run a script with the name name as the task
			script := fmt.Sprintf("%s/%s.sh", exec.Config.ScriptDir, task)
			err := exec.runCommand(script, taskConfig.Dir, env)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (exec *Executor) runCommand(cmd string, dir string, env []string) error {
	f, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return err
	}

	r, err := interp.New(
		interp.Params("-e"),
		interp.Env(expand.ListEnviron(env...)),
		interp.ExecHandler(interp.DefaultExecHandler(15*time.Second)),
		interp.OpenHandler(interp.DefaultOpenHandler()),
		interp.StdIO(exec.Stdin, exec.Stdout, exec.Stderr),
		interp.Dir(dir),
	)
	if err != nil {
		return err
	}

	err = r.Run(context.TODO(), f)
	if err != nil {
		return err
	}
	return nil
}

func (exec *Executor) ListTasksFromTaskFile(config *Config) {
	indent := "  "

	alphaTaskList := alphabetizeTaskList(&config.Tasks)
	for _, task := range *alphaTaskList {
		fmt.Printf("[%s]\n", task)

		if len(config.Tasks[task].Deps) > 0 {
			out := "\n"
			for _, depGroup := range config.Tasks[task].Deps {
				out += fmt.Sprintf("%s%s\n", indent+indent, depGroup)
			}
			fmt.Printf("%sdeps: %s", indent, out)
		}

		if len(config.Tasks[task].Cmds) > 0 {
			out := "\n"
			for _, cmd := range config.Tasks[task].Cmds {
				out += fmt.Sprintf("%s\"%s\"\n", indent+indent, cmd)
			}
			fmt.Printf("%scmds: %s\n", indent, out)
		} else {
			out := fmt.Sprintf("script: %s/%s.sh\n\n", exec.Config.ScriptDir, task)
			fmt.Printf("%s%s", indent, out)
		}
	}
}

func NewTaskConfig(taskFile string) (*Config, error) {
	var err error
	if taskFile == "" {
		dir, _ := os.Getwd()
		taskFile, err = findTaskFile(dir, "tasks.toml")
		if err != nil {
			return nil, err
		}
	}

	// parse the task file
	var config Config
	if _, err := toml.DecodeFile(taskFile, &config); err != nil {
		return nil, err
	}

	// set the task file dir, used as the base for a task's working directory
	config.TaskFileDir = filepath.Dir(taskFile)

	// set the script dir
	config.ScriptDir = "scripts"

	return &config, nil
}

func findTaskFile(dir, taskFile string) (string, error) {
	path := filepath.Join(dir, taskFile)

	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	parent := filepath.Dir(dir)
	if parent == dir {
		return "", fmt.Errorf("could not locate tasks.toml in current directory or parents")
	}

	return findTaskFile(parent, taskFile)
}
