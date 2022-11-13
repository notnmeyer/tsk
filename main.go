package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/naoina/toml"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type taskConfig struct {
	Tasks map[string]Tasks
}

type Tasks struct {
	Cmds []string
	Dir  string
	Env  map[string]string
}

type Opts struct {
	Stdout io.Writer
	Stdin  io.Reader
	Stderr io.Writer
}

func main() {
	f, err := os.Open("task.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var config taskConfig
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}

	for name, task := range config.Tasks {
		fmt.Printf("[%s]\n", name)

		if task.Dir == "" {
			task.Dir = "."
		}

		var env []string
		if len(task.Env) > 0 {
			env = ConvertEnv(task.Env)
		} else {
			env = os.Environ()
		}

		opts := Opts{
			Stdout: os.Stdout,
			Stdin:  os.Stdin,
			Stderr: os.Stderr,
		}

		// if a task contains tasks, run them
		if len(task.Cmds) > 0 {
			for _, cmd := range task.Cmds {
				err = RunCommand(cmd, task.Dir, env, opts)
				if err != nil {
					panic(err)
				}
			}
			// if there are no cmds, assume we intend to run a script with the name name as the task
		} else {
			script := fmt.Sprintf("./scripts/%s.sh", name)
			err = RunCommand(script, task.Dir, env, opts)
			if err != nil {
				panic(err)
			}
		}
	}
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

func ConvertEnv(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}
