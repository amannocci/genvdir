# Genvdir

Re-implementation of the daemontool's envdir written from Python to Golang.

like envdir, genvdir runs another program with environment modified according to files in a specified directory.

## Interface
```
   genvdir dir prog... [flags]
```
`dir` is a single argument.
`prog` consist of one or more arguments.

## Behavior
genvdir sets various environment variables as specified by files in the directory named `dir`. It then runs `prog`.

If `dir` contains a file named `s` whose first line is `t`, envdir removes an environment variable named `s` if one exists, and then adds an environment variable named `s` with value `t`. The name `s` must *not* contain `=`. Spaces and tabs at the end of `t` are removed. *Nulls* in `t` are changed to newlines in the environment variable.

If the file `s` is completely empty (0 bytes long), envdir removes an environment variable named `s` if one exists, without adding a new variable.

envdir exits **111** if it has trouble reading `dir`, if it runs out of memory for environment variables, or if it cannot run child. Otherwise its exit code is the same as that of child.
