package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoina/toml"
	flag "github.com/spf13/pflag"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type taskConfig struct {
	Tasks     map[string]Task
	ScriptDir string
}

type Task struct {
	Cmds []string
	Dir  string
	Env  map[string]string
}

type Executor struct {
	Stdout io.Writer
	Stdin  io.Reader
	Stderr io.Writer
	Config *taskConfig
}

func main() {
	var (
		listTasks bool
		taskFile  string
		tasks     []string
	)

	flag.BoolVarP(&listTasks, "list", "l", false, "list tasks")
	flag.StringVarP(&taskFile, "file", "f", "task.toml", "taskfile to use")
	flag.Parse()
	tasks = flag.Args()

	cfg, err := NewTaskConfig(taskFile)
	if err != nil {
		panic(err)
	}

	exec := Executor{
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
		Config: cfg,
	}

	// verify the tasks provided at the command line exist in the task file
	err = verifyTasks(exec.Config, tasks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if listTasks {
		exec.listTasksFromTaskFile(exec.Config)
	} else {
		err = exec.runTasks(exec.Config, &tasks)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func NewTaskConfig(taskFile string) (*taskConfig, error) {
	// open the task file
	f, err := os.Open(taskFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// parse the task file
	var config taskConfig
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	// set the script dir
	config.ScriptDir = scriptDir(taskFile)

	return &config, nil
}

func (exec *Executor) runTasks(config *taskConfig, tasks *[]string) error {
	for _, task := range *tasks {
		fmt.Printf("-- Task [%s]\n", task)

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
				err := exec.RunCommand(cmd, taskConfig.Dir, env)
				if err != nil {
					panic(err)
				}
			}
			// if there are no cmds then we intend to run a script with the name name as the task
		} else {
			script := fmt.Sprintf("%s/%s.sh", exec.Config.ScriptDir, task)
			err := exec.RunCommand(script, taskConfig.Dir, env)
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
}

// returns the expected script dir, which is "scripts" relative to the task file
func scriptDir(taskFile string) string {
	return fmt.Sprintf("%s/scripts", filepath.Dir(taskFile))
}

func (exec *Executor) RunCommand(cmd string, dir string, env []string) error {
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

func ConvertEnvToStringSlice(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}

// verifies the tasks provided at the command line exist
func verifyTasks(config *taskConfig, tasks []string) error {
	for _, task := range tasks {
		if _, ok := config.Tasks[task]; !ok {
			return fmt.Errorf("task '%s' not found in taskfile", task)
		}
	}
	return nil
}

func (exec *Executor) listTasksFromTaskFile(config *taskConfig) {
	for task := range config.Tasks {
		fmt.Println(task)
		if len(config.Tasks[task].Cmds) > 0 {
			fmt.Printf("	cmds: %v\n", config.Tasks[task].Cmds)
		} else {
			fmt.Printf("	script: %s/%s.sh\n", exec.Config.ScriptDir, task)
		}
	}
}
