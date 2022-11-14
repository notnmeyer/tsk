package main

import (
	"fmt"
	"os"

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

	if err := verifyTasks(exec.Config, tasks); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if listTasks {
		exec.ListTasksFromTaskFile(exec.Config)
	} else {
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
	}
	return nil
}
