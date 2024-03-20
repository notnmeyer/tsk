<p align="center">
  <img src="https://github.com/notnmeyer/tsk/assets/672246/bb8c8c55-756a-4eb0-be9f-9697267bb012" width="20%" />
  <p align="center"><b>tsk</b>: a simple task runner with TOML syntax</p>
</p>

## example

```
$ cat tasks.toml
[tasks.hello_world]
cmds = ["echo Hello World!"]

$ tsk hello_world
Hello World!
```

see `examples/tasks.toml` for complete usage and configuration reference.

## installation

via [Homebrew](https://brew.sh):

```
brew tap notnmeyer/tsk
brew install tsk
```

alternatively, you can install directly from [releases](https://github.com/notnmeyer/tsk/releases) page or see the [install_release](./scripts/install_release.sh) script.

### from source

- **without tsk installed** `go build -o ./bin/tsk -v cmd/tsk/tsk.go`
- **with tsk installed** `tsk build`

## features

### tasks.toml locations

tsk will look for a `tasks.toml` file in the current directory, looking in parent directories if one isn't found.

you can specify a file in another location or with another name with the `--file` flag.

### dependencies and dependency groups

tasks can depend on other tasks as dependencies via a task's `deps` key. dependencies are organized in groups represented in toml as nested arrays. dependency groups are executed in the order they are defined in the `deps` key, while tasks within a group are executed in parallel.

```
[tasks.one]
cmds = ["echo one"]

[tasks.two]
cmds = ["echo two"]

[tasks.three]
cmds = ["echo three"]

[tasks.main]
deps = [
  ["one", "two"], # one and two will run in parallel
  ["three"],      # three will run after one and two have finished
]
cmds = ["echo main"]
```

### let shell be shell

writing shell in toml or yaml files sucks—you miss out of syntax highlighting, linting, and other tools like shellcheck.

you may omit a task's `cmds` field, which instead runs a script with the same name as the task from the `tsk` directory relative to the location of your `tasks.toml` file.

if you need to write anything more complicated than one or two short commands in a task, use a script!

try `tsk --file examples/tasks.toml no_cmd` to see this in action.

### environment variables

tsk loads environment variables and merges them with the precendence listed below. items earlier in the list are overriden by items lower in the list.

1. the top-level `env` key
1. the top-level `dotenv` key
1. `tasks.<task_name>.env`
1. the parent process, e.g., `MY_VAR=hey tsk ...`
1. `tasks.<task_name>.dotenv`

#### "pure" tasks

setting `pure = true` on a task will prevent the parent process's environment from being inherited. Similarly, the CLI argument `--pure` will prevent _any_ task from inheriting the parent's env. the only exceptions are `$USER` and `$HOME`, which are always inherited.

### toml 1.1.0

some [upcoming toml 1.1.0 features are supported](https://github.com/toml-lang/toml/blob/main/CHANGELOG.md#unreleased). of note, newlines and trailing commas are now allowed in inline tables:

```toml
[tasks.example]
env = {
  msg1 = "hello",
  msg2 = "world",
}
cmds = ["echo $msg1 $msg2"]
```
