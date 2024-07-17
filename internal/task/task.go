package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	output "github.com/notnmeyer/tsk/internal/outputformat"

	"github.com/BurntSushi/toml"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// represents parsed task file
type Config struct {
	DotEnv      string            `toml:"dotenv"`
	Env         map[string]string `toml:"env"`
	Tasks       map[string]Task   `toml:"tasks"`
	ScriptDir   string            `toml:"script_dir"`
	TaskFileDir string
}

// represents an individual task
type Task struct {
	Cmds        []string          `toml:"cmds"`
	Deps        [][]string        `toml:"deps"`
	Description string            `toml:"description"`
	Dir         string            `toml:"dir"`
	Env         map[string]string `toml:"env"`
	DotEnv      string            `toml:"dotenv"`
	Pure        bool              `toml:"pure"`
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
					fmt.Println(err.Error())
					// if the cmd exited with an error, bail immediately
					os.Exit(1)
				}
			}
		} else {
			// if there are no cmds then we intend to run a script with the name name as the task
			script := fmt.Sprintf("%s/%s", exec.Config.ScriptDir, task)
			err := exec.runCommand(script, taskConfig.Dir, env)
			if err != nil {
				fmt.Println(err.Error())
				// if the cmd exited with an error, bail immediately
				os.Exit(1)
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

func (exec *Executor) ListTasksFromTaskFile(regex *regexp.Regexp, format output.OutputFormat) {
	tasks := filterTasks(&exec.Config.Tasks, regex)
	indent := "  "

	switch format {
	case output.Markdown:
		for name, t := range tasks {
			fmt.Printf("## %s\n", name)
			if len(t.Cmds) > 0 {
				for _, cmd := range t.Cmds {
					fmt.Printf("%s- %s\n", indent, cmd)
				}
			} else {
				fmt.Printf("%s- %s/%s\n", indent, exec.Config.ScriptDir, name)
			}
		}
	case output.TOML:
		toml.NewEncoder(os.Stdout).Encode(tasks)
	case output.Text:
		for name, t := range tasks {
			// name
			fmt.Printf("%s:\n", name)

			// description
			if t.Description != "" {
				fmt.Printf("%sdescription:\n", indent)
				trimmed := strings.TrimSpace(t.Description)
				for _, line := range strings.Split(trimmed, "\n") {
					fmt.Printf("%s%s\n", strings.Repeat(indent, 2), line)
				}
			}

			// deps
			if len(t.Deps) > 0 {
				fmt.Printf("%sdeps:\n", indent)
				for _, dep := range t.Deps {
					fmt.Printf("%s%v\n", indent+indent, dep)
				}
			}

			// cmds
			fmt.Printf("%scommands:\n", indent)
			if len(t.Cmds) > 0 {
				for _, cmd := range t.Cmds {
					fmt.Printf("%s\n", indent+indent+cmd)
				}
			} else {
				fmt.Printf("%s%s/%s\n", indent+indent, exec.Config.ScriptDir, name)
			}

			// dir
			if t.Dir != "" {
				fmt.Printf("%sdir: %s\n", indent, t.Dir)
			}

			// dotenv
			if t.DotEnv != "" {
				fmt.Printf("%sdotenv: %s\n", indent, t.DotEnv)
			}

			// pure
			if t.Pure == true {
				fmt.Printf("%spure: %t\n", indent, t.Pure)
			}

			fmt.Println("")
		}
	}
}

func filterTasks(tasks *map[string]Task, regex *regexp.Regexp) map[string]Task {
	filtered := make(map[string]Task)
	for k, v := range *tasks {
		if regex.MatchString(k) {
			filtered[k] = v
		}
	}
	return filtered
}

func NewTaskConfig(taskFile, cliArgs string) (*Config, error) {
	var err error
	if taskFile == "" {
		dir, _ := os.Getwd()
		taskFile, err = findTaskFile(dir, "tasks.toml")
		if err != nil {
			return nil, err
		}
	}

	// render the task file as a template
	rendered, err := render(taskFile, cliArgs)
	if err != nil {
		return nil, err
	}

	// parse the task file
	var config Config
	if _, err := toml.Decode(rendered.String(), &config); err != nil {
		return nil, err
	}

	// set the task file dir, used as the base for a task's working directory
	config.TaskFileDir = filepath.Dir(taskFile)

	// set the script dir
	if len(config.ScriptDir) == 0 {
		config.ScriptDir = "tsk"
	}

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

func InitTaskfile() error {
	content := `[tasks.hello]
cmds = ["echo hello!"]
`

	cwd, _ := os.Getwd()
	filePath := path.Join(cwd, "tasks.toml")

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("%s exists", filePath)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}
