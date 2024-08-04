# check-changes

*A basic tool for warning you about things you're checking in.*

## Basic outline

check-changes will by default look at staged changes in the current index (via git). By passing a ref name, it will diff against that ref (say, if you have a branch you're looking to clean up before making a PR). Major checks return status code 1 to make this program suitable for use in a git hook.

Major checks (will return status code 1):

- NOCHECKIN: if this string appears anywhere in added lines
- whitespace: if any added lines have leading whitespace which differs from the first detected leading whitespace in that file

Lesser checks (will print output but return status code 0):

- TODO: if this string appears anywhere in added lines
- stash entries: if any entries in `git stash list` contain the current branch name, which may indicate that the user forgot some changes they had previously stashed

## Building the project

### Requirements:

- a Golang installation (built & tested on Go v1.22)
- an internet connection to download dependencies (only necessary if dependencies have changed or this is the first build)
- a `make` installation. This project is built with GNU make v4; full compatibility with other versions of make (such as that shipped by Apple) is not guaranteed, but it _should_ be broadly compatible.

To build the project, simply run `make` in the project's root directory to build the output executable.

> _Note: running with `make` is not strictly necessary. Reference the provided `Makefile` for typical development commands._