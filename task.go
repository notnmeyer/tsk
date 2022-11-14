package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoina/toml"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// represents parsed task file
type Config struct {
	Tasks     map[string]Task
	ScriptDir string
}

// represents an individual task
type Task struct {
	Cmds []string
	Dir  string
	Env  map[string]string
}

type Executor struct {
	Stdout io.Writer
	Stdin  io.Reader
	Stderr io.Writer
	Config *Config
}

func (exec *Executor) RunTasks(config *Config, tasks *[]string) error {
	for _, task := range *tasks {
		fmt.Printf("-- task [%s]\n", task)

		taskConfig := config.Tasks[task]
		if taskConfig.Dir == "" {
			taskConfig.Dir = "."
		}

		var env []string
		if len(taskConfig.Env) > 0 {
			env = ConvertEnvToStringSlice(taskConfig.Env)
		} else {
			env = os.Environ()
		}

		// if a task contains cmds, run them
		if len(taskConfig.Cmds) > 0 {
			for _, cmd := range taskConfig.Cmds {
				err := exec.runCommand(cmd, taskConfig.Dir, env)
				if err != nil {
					panic(err)
				}
			}
			// if there are no cmds then we intend to run a script with the name name as the task
		} else {
			script := fmt.Sprintf("%s/%s.sh", exec.Config.ScriptDir, task)
			err := exec.runCommand(script, taskConfig.Dir, env)
			if err != nil {
				panic(err)
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
	for task := range config.Tasks {
		fmt.Println(task)
		if len(config.Tasks[task].Cmds) > 0 {
			fmt.Printf("	cmds: %v\n", config.Tasks[task].Cmds)
		} else {
			fmt.Printf("	script: %s/%s.sh\n", exec.Config.ScriptDir, task)
		}
	}
}

func NewTaskConfig(taskFile string) (*Config, error) {
	// open the task file
	f, err := os.Open(taskFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// parse the task file
	var config Config
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	// set the script dir
	config.ScriptDir = scriptDir(taskFile)

	return &config, nil
}

// returns the expected script dir, which is "scripts" relative to the task file
func scriptDir(taskFile string) string {
	return fmt.Sprintf("%s/scripts", filepath.Dir(taskFile))
}

func ConvertEnvToStringSlice(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}
