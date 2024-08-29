package main

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
	flag.StringVarP(&opts.output, "output", "o", "text", fmt.Sprintf("output format (applies only to --list) (one of: %s, %s)", string(output.Text), string(output.Markdown)))
	flag.BoolVarP(&opts.pure, "pure", "", false, "don't inherit the parent env")
	flag.StringVarP(&opts.taskFile, "file", "f", "", "taskfile to use")
	flag.BoolVarP(&help, "help", "h", false, "")
	flag.Parse()

	if help {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if opts.init {
		if err := task.InitTaskfile(); err != nil {
			fmt.Printf("couldn't init: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("created tasks.toml!\n")
		return
	}

	if opts.displayVersion {
		fmt.Printf("tsk v%s, git:%s\n", version, commit)
		return
	}

	if !output.IsValid(opts.output) {
		fmt.Printf("--output must one of: %s\n", output.String())
		os.Exit(1)
	}

	// check if there are args passed after "--".
	//   - if "--" is not present ArgsLenAtDash() returns -1.
	//   - dash position 0 would be invocations like, `tsk -l -- foo`
	if flag.CommandLine.ArgsLenAtDash() >= 0 {
		opts.tasks = flag.Args()[:flag.CommandLine.ArgsLenAtDash()]
		opts.cliArgs = strings.Join(flag.Args()[flag.CommandLine.ArgsLenAtDash():], " ")
	} else {
		opts.tasks = flag.Args()
	}

	// cfg is the parsed task file
	cfg, err := task.NewTaskConfig(opts.taskFile, opts.cliArgs)
	if err != nil {
		panic(err)
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
	if err := verifyTasks(exec.Config, opts.tasks); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exec.RunTasks(exec.Config, &opts.tasks)
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
