<div align="center">
  <h1>tsk</h1>
  <p>A task-runner and build tool for simpletons</p>
  https://notnmeyer.github.io/tsk-docs/
</div>

## installation

https://notnmeyer.github.io/tsk-docs/docs/installation

tl;dr, `brew install notnmeyer/tsk/tsk`

## example

```
$ cat tasks.toml
[tasks.hello_world]
cmds = ["echo Hello World!"]

$ tsk hello_world
Hello World!
```

see `examples/tasks.toml` for complete usage and configuration reference.

## docs

https://notnmeyer.github.io/tsk-docs/

## release

tag a new release with, `tsk release -- v0.0.0`, and let GHA do it its thing.
