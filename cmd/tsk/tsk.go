package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/notnmeyer/tsk/internal/task"
	flag "github.com/spf13/pflag"
)

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

	// verify the tasks at the cli exist
	if err := verifyTasks(exec.Config, tasks); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if listTasks {
		exec.ListTasksFromTaskFile(exec.Config)
		return
	}

	var wg sync.WaitGroup
	for _, task := range tasks {

		// TODO: this doesn't run deps of deps

		// run the task's dependencies
		if len(exec.Config.Tasks[task].Deps) > 0 {
			wg.Add(len(exec.Config.Tasks[task].Deps))
			for _, dep := range exec.Config.Tasks[task].Deps {
				go func(dep string) {
					defer wg.Done()
					// TODO: capture errors and return them through a channel
					exec.RunTasks(exec.Config, &[]string{dep})
				}(dep)
			}
			wg.Wait()
		}

		// run the primary task
		if err := exec.RunTasks(exec.Config, &tasks); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
			for _, dep := range config.Tasks[task].Deps {
				if _, ok := config.Tasks[dep]; !ok {
					return fmt.Errorf("task '%s' not found in taskfile", dep)
				}
			}
		}
	}
	return nil
}
