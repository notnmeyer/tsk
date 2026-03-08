package task

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"text/template"

	"github.com/joho/godotenv"
)

// ANSI color codes for prefix coloring. These are chosen for visual distinction.
var prefixColors = []string{
	"\033[36m",  // cyan
	"\033[33m",  // yellow
	"\033[35m",  // magenta
	"\033[32m",  // green
	"\033[34m",  // blue
	"\033[31m",  // red
	"\033[96m",  // bright cyan
	"\033[93m",  // bright yellow
	"\033[95m",  // bright magenta
	"\033[92m",  // bright green
	"\033[94m",  // bright blue
	"\033[91m",  // bright red
}

const colorReset = "\033[0m"

type prefixWriter struct {
	writer io.Writer
	prefix string
	buf    []byte
}

func newPrefixWriter(w io.Writer, prefix string, color string) *prefixWriter {
	coloredPrefix := color + prefix + colorReset
	return &prefixWriter{writer: w, prefix: coloredPrefix}
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	for {
		idx := bytes.IndexByte(p.buf, '\n')
		if idx < 0 {
			break
		}
		line := p.buf[:idx+1]
		if _, err := fmt.Fprintf(p.writer, "%s%s", p.prefix, line); err != nil {
			return 0, err
		}
		p.buf = p.buf[idx+1:]
	}
	return len(b), nil
}

type Vals struct {
	CLI_ARGS string
}

func alphabetizeTaskList(t *map[string]Task) *[]string {
	var taskNames []string
	for taskName := range *t {
		taskNames = append(taskNames, taskName)
	}
	sort.Strings(taskNames)
	return &taskNames
}

func ConvertEnvToStringSlice(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}

func readDotEnv(filename string) (map[string]string, error) {
	dotEnv, err := godotenv.Read(filename)
	if err != nil {
		return nil, err
	}

	return dotEnv, nil
}

func appendDotEnvToEnv(env []string, dotenv string) ([]string, error) {
	additionalEnv, err := readDotEnv(dotenv)
	if err != nil {
		// the dotenv file missing is non-fatal. log a warning and continue
		fmt.Printf("Warning: Could not load dotenv file %s: %v\n", dotenv, err)
		return env, nil
	}
	env = append(env, ConvertEnvToStringSlice(additionalEnv)...)
	return env, nil
}

func render(file, cliArgs string, cliArgsPlaceholder bool) (*bytes.Buffer, error) {
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		return nil, err
	}

	// insert a placeholder value for cliArgs for display purposes
	if cliArgsPlaceholder && cliArgs == "" {
		cliArgs = "{{.CLI_ARGS}}"
	}

	var renderedBuffer bytes.Buffer
	if err := tmpl.Execute(&renderedBuffer, &Vals{CLI_ARGS: cliArgs}); err != nil {
		return nil, err
	}

	return &renderedBuffer, nil
}
