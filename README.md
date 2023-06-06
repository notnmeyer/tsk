# tsk
a simple task runner. supports TOML 1.1.0!

```
$ cat tasks.toml
[tasks.hello_world]
cmds = ["echo Hello World!"]

$ tsk hello_world
Hello World!
```

see `examples/tasks.toml` for complete usage and configuration reference.

## installation

### prebuilt binary
1. head over to the [releases](https://github.com/notnmeyer/tsk/releases) and grab the URL for your desired release
1. download it, `wget https://github.com/notnmeyer/tsk/releases/download/v0.1.0/tsk_v0.1.0_Darwin_arm64.tar.gz`
1. extract it, `tar -xzf /tmp/tsk_v0.1.0_Darwin_arm64.tar.gz`
1. move it to somewhere in your $PATH, `mv ./tsk ~/bin/`

```
➜ tsk --version
tsk v0.1.0, git:46e7d24edc54d07b38c476b167a79a46c091160b
```

### from source
1. clone this repo
1. install goreleaser
1. tsk uses tsk as its task runner, so if you're bootstrapping, take a look in `tasks.toml` and run the build command manually. you should wind up with a binary at `./bin/tsk`. then you can use the binary to build itself going forward.


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

you may omit a task's `cmds` field, which instead runs a script with the same name as the task from the `scripts` directory relative to the location of your `tasks.toml` file.

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
