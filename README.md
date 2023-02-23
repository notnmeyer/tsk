# tsk
a simple task runner.

```
$ cat tasks.toml
[tasks.hello_world]
cmds = ["echo Hello World!"]

$ tsk hello_world
Hello World!
```

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
