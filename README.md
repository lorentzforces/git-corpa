# git-corpa

*Custom extensions for day-to-day git usage at the corpa*

## Background

This project was originally known as `check-changes`, and was built to do some basic sanity checking of current changes (or changes on a branch) before pushing code. I came up with some other git-related conveniences I'd like to have, so it seemed natural to package all of them together into a suite of convenience tools.

These extensions are built by and for someone doing pretty typical enterprise development on git repositories. Many of the functions are going to assume a shared central repository as a source of truth, and will operate accordingly. Don't @ me bro.

## Basic outline

check-changes will by default look at staged changes in the current index (via git). By passing a ref name, it will diff against that ref (say, if you have a branch you're looking to clean up before making a PR). Major checks return status code 1 to make this program suitable for use in a git hook.

Major checks (will return status code 1):

- NOCHECKIN: if this string appears anywhere in added lines
- whitespace: if any added lines have leading whitespace which differs from the first detected leading whitespace in that file

Lesser checks (will print output but return status code 0):

- TODO: if this string appears anywhere in added lines
- stash entries: if any entries in `git stash list` contain the current branch name, which may indicate that the user forgot some changes they had previously stashed

## Run requirements

- A `git` executable available somewhere on your system `PATH`.

## Building the project

### Build requirements:

- a Golang installation (built & tested on Go v1.22)
- an internet connection to download dependencies (only necessary if dependencies have changed or this is the first build)
- a `make` installation. This project is built with GNU make v4; full compatibility with other versions of make (such as that shipped by Apple) is not guaranteed, but it _should_ be broadly compatible.

To build the project, simply run `make` in the project's root directory to build the output executable.

> _Note: running with `make` is not strictly necessary. Reference the provided `Makefile` for typical development commands._

## Current project to-dos (in no particular order):

- Use this a bit to shake down bugs.
- Colorize CLI output (when stdout is not a pipe).
- Make checked keyword searching configurable so people can set their own custom keywords.
- Rejigger this into an executable with subcommands.
- Add a smart branch create (start at main, auto-create a remote tracking branch)
- Add a smart branch delete (check if already orphaned, confirm deletion, smart force-delete)
