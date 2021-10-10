# genvdir
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

*Re-implementation of the daemontool's envdir written from Python to Golang.
like envdir, genvdir runs another program with environment modified according to files in a specified directory.*
* [Source](https://github.com/amannocci/genvdir)
* [Issues](https://github.com/amannocci/genvdir/issues)
* [Contact](mailto:adrien.mannocci@gmail.com)

## Prerequisites
* [Golang](https://golang.org/) for development.

## Usage

### How to use genvdir

```
   genvdir dir prog... [flags]
```
`dir` is a single argument.
`prog` consist of one or more arguments.

### Behavior

genvdir sets various environment variables as specified by files in the directory named `dir`. It then runs `prog`.

If `dir` contains a file named `s` whose first line is `t`, envdir removes an environment variable named `s` if one exists, and then adds an environment variable named `s` with value `t`. The name `s` must *not* contain `=`. Spaces and tabs at the end of `t` are removed. *Nulls* in `t` are changed to newlines in the environment variable.

If the file `s` is completely empty (0 bytes long), envdir removes an environment variable named `s` if one exists, without adding a new variable.

envdir exits **111** if it has trouble reading `dir`, if it runs out of memory for environment variables, or if it cannot run child. Otherwise its exit code is the same as that of child.


## Contributing
If you find this project useful here's how you can help :

* Send a Pull Request with your awesome new features and bug fixed
* Be a part of the ommunity and help resolve [Issues](https://github.com/amannocci/genvdir/issues)
