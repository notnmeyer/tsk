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
	Tasks map[string]Task
}

type Task struct {
	Cmds []string
	Dir  string
	Env  map[string]string
}

type Opts struct {
	Stdout io.Writer
	Stdin  io.Reader
	Stderr io.Writer
}

var (
	cliTasks  []string // the tasks to run from the command line
	listTasks bool     // if true, list the tasks in the task file and dont run any tasks
	taskFile  string   // the task file to use
)

func main() {
	parseFlags()

	// open the task file
	f, err := os.Open(taskFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// parse the task file
	var config taskConfig
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}

	// verify the tasks provided at the command line exist in the task file
	err = verifyTasks(&config, cliTasks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if listTasks {
		listTasksFromTaskFile(&config)
	} else {
		err = runTasks(&config, &cliTasks)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func runTasks(config *taskConfig, tasks *[]string) error {
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

		opts := Opts{
			Stdout: os.Stdout,
			Stdin:  os.Stdin,
			Stderr: os.Stderr,
		}

		// if a task contains cmds, run them
		if len(taskConfig.Cmds) > 0 {
			for _, cmd := range taskConfig.Cmds {
				err := RunCommand(cmd, taskConfig.Dir, env, opts)
				if err != nil {
					panic(err)
				}
			}
			// if there are no cmds then we intend to run a script with the name name as the task
		} else {
			script := fmt.Sprintf("%s/%s.sh", scriptDir(), task)
			err := RunCommand(script, taskConfig.Dir, env, opts)
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
}

// returns the expected script dir, which is "scripts" relative to the task file
func scriptDir() string {
	return fmt.Sprintf("%s/scripts", filepath.Dir(taskFile))
}

func RunCommand(cmd string, dir string, env []string, opts Opts) error {
	f, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return err
	}

	r, err := interp.New(
		interp.Params("-e"),
		interp.Env(expand.ListEnviron(env...)),
		interp.ExecHandler(interp.DefaultExecHandler(15*time.Second)),
		interp.OpenHandler(interp.DefaultOpenHandler()),
		interp.StdIO(opts.Stdin, opts.Stdout, opts.Stderr),
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

func parseFlags() {
	flag.StringVarP(&taskFile, "file", "f", "task.toml", "taskfile to use")
	flag.BoolVarP(&listTasks, "list", "l", false, "list tasks")
	flag.Parse()
	cliTasks = flag.Args()
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

func listTasksFromTaskFile(config *taskConfig) {
	for task := range config.Tasks {
		fmt.Println(task)
		if len(config.Tasks[task].Cmds) > 0 {
			fmt.Printf("	cmds: %v\n", config.Tasks[task].Cmds)
		} else {
			fmt.Printf("	script: %s/%s.sh\n", scriptDir(), task)
		}
	}
}
