package main

import (
	"fmt"
	"os"

	"github.com/notnmeyer/tsk/internal/task"
	flag "github.com/spf13/pflag"
)

var version, commit string

func init() {
	// TOML 1.1 features are behind a flag until officially released
	os.Setenv("BURNTSUSHI_TOML_110", "")
}

func main() {
	var (
		displayVersion bool
		listTasks      bool
		pure           bool
		taskFile       string
		tasks          []string
	)

	flag.BoolVarP(&displayVersion, "version", "V", false, "display tsk version")
	flag.BoolVarP(&listTasks, "list", "l", false, "list tasks")
	flag.BoolVarP(&pure, "pure", "", false, "don't inherit the parent env")
	flag.StringVarP(&taskFile, "file", "f", "", "taskfile to use")
	flag.Parse()
	tasks = flag.Args()

	if displayVersion {
		fmt.Printf("tsk v%s, git:%s\n", version, commit)
		return
	}

	// cfg is the parsed task file
	cfg, err := task.NewTaskConfig(taskFile)
	if err != nil {
		panic(err)
	}

	exec := task.Executor{
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
		Config: cfg,
	}

	if listTasks {
		exec.ListTasksFromTaskFile(exec.Config)
		return
	}

	if pure {
		for name, task := range exec.Config.Tasks {
			task.Pure = true
			exec.Config.Tasks[name] = task
		}
	}

	// verify the tasks at the cli exist
	if err := verifyTasks(exec.Config, tasks); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := exec.RunTasks(exec.Config, &tasks); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// verifies the tasks provided at the command line exist
func verifyTasks(config *task.Config, tasks []string) error {
	for _, task := range tasks {
		if _, ok := config.Tasks[task]; !ok {
			return fmt.Errorf("task '%s' not found in taskfile", task)
		}

		// if a task specifies deps, verify they exist
		if len(config.Tasks[task].Deps) > 0 {
			for _, depGroup := range config.Tasks[task].Deps {
				for _, dep := range depGroup {
					if _, ok := config.Tasks[dep]; !ok {
						return fmt.Errorf("task '%s' not found in taskfile", dep)
					}
				}
			}
		}
	}
	return nil
}
