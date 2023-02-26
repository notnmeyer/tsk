package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// represents parsed task file
type Config struct {
	Tasks       map[string]Task
	ScriptDir   string
	TaskFileDir string
}

// represents an individual task
type Task struct {
	Cmds   []string
	Deps   [][]string
	Dir    string
	Env    map[string]string
	DotEnv string
}

type Executor struct {
	Stdout io.Writer
	Stdin  io.Reader
	Stderr io.Writer
	Config *Config
}

func (exec *Executor) RunTasks(config *Config, tasks *[]string) error {
	for _, task := range *tasks {
		taskConfig := config.Tasks[task]

		if taskConfig.Dir == "" {
			taskConfig.Dir = config.TaskFileDir
		}

		var env []string
		if len(taskConfig.Env) > 0 {
			env = ConvertEnvToStringSlice(taskConfig.Env)
		} else {
			env = os.Environ()
		}

		// append dotenv
		if taskConfig.DotEnv != "" {
			dotEnv, err := godotenv.Read(filepath.Join(taskConfig.Dir, taskConfig.DotEnv))
			if err != nil {
				return err
			}

			env = append(env, ConvertEnvToStringSlice(dotEnv)...)
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

		// if a task contains cmds, run them
		if len(taskConfig.Cmds) > 0 {
			for _, cmd := range taskConfig.Cmds {
				err := exec.runCommand(cmd, taskConfig.Dir, env)
				if err != nil {
					return err
				}
			}
		} else {
			// if there are no cmds then we intend to run a script with the name name as the task
			script := fmt.Sprintf("%s/%s.sh", exec.Config.ScriptDir, task)
			err := exec.runCommand(script, taskConfig.Dir, env)
			if err != nil {
				return err
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
	indent := "  "

	alphaTaskList := alphabetizeTaskList(&config.Tasks)
	for _, task := range *alphaTaskList {
		fmt.Printf("[%s]\n", task)

		if len(config.Tasks[task].Deps) > 0 {
			out := "\n"
			for _, depGroup := range config.Tasks[task].Deps {
				out += fmt.Sprintf("%s%s\n", indent+indent, depGroup)
			}
			fmt.Printf("%sdeps: %s", indent, out)
		}

		if len(config.Tasks[task].Cmds) > 0 {
			out := "\n"
			for _, cmd := range config.Tasks[task].Cmds {
				out += fmt.Sprintf("%s\"%s\"\n", indent+indent, cmd)
			}
			fmt.Printf("%scmds: %s\n", indent, out)
		} else {
			out := fmt.Sprintf("script: %s/%s.sh\n\n", exec.Config.ScriptDir, task)
			fmt.Printf("%s%s", indent, out)
		}
	}
}

func NewTaskConfig(taskFile string) (*Config, error) {
	var err error
	if taskFile == "" {
		dir, _ := os.Getwd()
		taskFile, err = findTaskFile(dir, "tasks.toml")
		if err != nil {
			return nil, err
		}
	}

	// parse the task file
	var config Config
	if _, err := toml.DecodeFile(taskFile, &config); err != nil {
		return nil, err
	}

	// set the task file dir, used as the base for a task's working directory
	config.TaskFileDir = filepath.Dir(taskFile)

	// set the script dir
	config.ScriptDir = "scripts"

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

func ConvertEnvToStringSlice(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}
