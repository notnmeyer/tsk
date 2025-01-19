package main

// Version and commit are set at build time via ldflags

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	output "github.com/notnmeyer/tsk/internal/outputformat"
	"github.com/notnmeyer/tsk/internal/task"

	flag "github.com/spf13/pflag"
)

var (
	version, commit string
	help            bool
)

type Options struct {
	cliArgs        string
	displayVersion bool
	filter         string
	init           bool
	listTasks      bool
	output         string
	pure           bool
	taskFile       string
	tasks          []string
	which          bool
}

const defaultOutputFormat = output.OutputFormat(output.Text)

func init() {
	// TOML 1.1 features are behind a flag until officially released
	os.Setenv("BURNTSUSHI_TOML_110", "")
}

func main() {
	opts := Options{}
	flag.BoolVarP(&opts.displayVersion, "version", "V", false, "display tsk version")
	flag.StringVarP(&opts.filter, "filter", "F", ".*", "regex filter for --list")
	flag.BoolVar(&opts.init, "init", false, "create a tasks.toml file in $PWD")
	flag.BoolVarP(&opts.listTasks, "list", "l", false, "list tasks")
	flag.StringVarP(&opts.output, "output", "o", "text", fmt.Sprintf("output format (applies only to --list) (one of: %s)", output.String()))
	flag.BoolVarP(&opts.pure, "pure", "", false, "don't inherit the parent env")
	flag.StringVarP(&opts.taskFile, "file", "f", "", "taskfile to use")
	flag.BoolVar(&opts.which, "which", false, "print the path to the found tasks.toml, or an error")
	flag.BoolVarP(&help, "help", "h", false, "")
	flag.Parse()

	// flags that exit early and don't require parsing the taskfile
	switch {
	case help:
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	case opts.displayVersion:
		fmt.Printf("tsk v%s, git:%s\n", version, commit)
		return
	case opts.init:
		if err := task.InitTaskfile(); err != nil {
			fmt.Printf("couldn't init: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("created tasks.toml!\n")
		return
	case !output.IsValid(opts.output):
		fmt.Printf("--output must one of: %s\n", output.String())
		os.Exit(1)
	}

	opts.tasks, opts.cliArgs = parseArgs(flag.Args())

	// cfg is the parsed task file
	cfg, err := task.NewTaskConfig(opts.taskFile, opts.cliArgs, opts.listTasks)
	if err != nil {
		panic(err)
	}

	if opts.which {
		fmt.Println(cfg.TaskFilePath)
	}

	exec := task.Executor{
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
		Config: cfg,
	}

	if opts.listTasks {
		exec.ListTasksFromTaskFile(regexp.MustCompile(opts.filter), output.OutputFormat(opts.output))
		return
	}

	if opts.pure {
		for name, task := range exec.Config.Tasks {
			task.Pure = true
			exec.Config.Tasks[name] = task
		}
	}

	// verify the tasks at the cli exist
	if err := exec.VerifyTasks(opts.tasks); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exec.RunTasks(exec.Config, &opts.tasks)
}

// splits args like [task1, task2 --, arg1, arg2] into
// - tasks = []string{"task1", "task2"}
// - cliArgs = "arg1 arg2"
func parseArgs(args []string) (tasks []string, cliArgs string) {
	cliArgsIndex := func() int {
		for index, arg := range args {
			if arg == "--" {
				return index
			}
		}
		return -1
	}()

	hasCliArgs := func() bool {
		if cliArgsIndex >= 0 {
			return true
		}
		return false
	}()

	if hasCliArgs {
		tasks = args[:cliArgsIndex]
		cliArgs = strings.Join(args[cliArgsIndex+1:], " ")
		return
	}

	return args, ""
}
