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
### dependencies
tasks can depend on other tasks as dependencies via a task's `deps` key. dependencies are run in parallel. a dependency may also have dependencies of its own.

### let shell be shell
writing shell in toml or yaml files sucks—you miss out of syntax highlighting, linting, and other tools like shellcheck.

you may omit a task's `cmds` field, which instead runs a script with the same name as the task from the `scripts` directory relative to the location of your `tasks.toml` file.

if you need to write anything more complicated than one or two short commands in a task, use a script!

try `tsk --file examples/tasks.toml no_cmd` to see this in action.
