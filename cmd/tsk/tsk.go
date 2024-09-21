package main

// Version and commit are set at build time via ldflags

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/notnmeyer/tsk/internal/openai"
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
	generate       bool
	init           bool
	listTasks      bool
	output         string
	pure           bool
	taskFile       string
	tasks          []string
	which          bool
}

const defaultOutputFormat = output.OutputFormat(output.Text)
const generateUsage = "tsk <name> --generate -- <description>"

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
	flag.BoolVarP(&opts.generate, "generate", "g", false, "use AI to generate a task with the name specified. pass the prompt for the task after '--'. usage: tsk fib --generate -- generate fib numbers up to 10")
	flag.BoolVarP(&help, "help", "h", false, "")
	flag.Parse()

	// options or commands that don't require parsing the tasks.toml and exit early
	switch {
	case help:
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	case opts.init:
		if err := task.InitTaskfile(); err != nil {
			fmt.Printf("couldn't init: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("created tasks.toml!\n")
		return
	case opts.displayVersion:
		fmt.Printf("tsk v%s, git:%s\n", version, commit)
		return
	case opts.generate:
		if len(opts.tasks) == 0 || opts.cliArgs == "" {
			fmt.Printf("you must describe the task to generate. usage: %s\n", generateUsage)
			os.Exit(1)
		}

		resp, err := openai.GenerateTask(opts.tasks[0], opts.cliArgs)
		if err != nil {
			panic(err)
		}

		fmt.Println(*resp)
		return
	}

	// separates the tasks to run any CLI_ARGS
	//
	// check if there are args passed after "--".
	//   - if "--" is not present ArgsLenAtDash() returns -1.
	//   - dash position 0 would be invocations like, `tsk -l -- foo`
	if flag.CommandLine.ArgsLenAtDash() >= 0 {
		opts.tasks = flag.Args()[:flag.CommandLine.ArgsLenAtDash()]
		opts.cliArgs = strings.Join(flag.Args()[flag.CommandLine.ArgsLenAtDash():], " ")
	} else {
		opts.tasks = flag.Args()
	}

	// parse the tasks.toml
	cfg, err := task.NewTaskConfig(opts.taskFile, opts.cliArgs, opts.listTasks)
	if err != nil {
		panic(err)
	}

	// the executor runs tasks
	exec := task.Executor{
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
		Config: cfg,
	}

	// display the complete path to the parsed tasks.toml
	if opts.which {
		fmt.Println(cfg.TaskFilePath)
	}

	if opts.listTasks {
		if !output.IsValid(opts.output) {
			fmt.Printf("--output must one of: %s\n", output.String())
			os.Exit(1)
		}
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
