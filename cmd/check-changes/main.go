package main

import (
	"fmt"
	"os"

	"github.com/lorentzforces/check-changes/internal/checking"
	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
	"github.com/spf13/pflag"
)

// NOTE: for now, this only checks staged changes
func main() {
	if !git.ExecExists() {
		platform.FailOut("\"git\" executable not found on system PATH")
	}

	var helpRequested bool
	pflag.BoolVarP(
		&helpRequested,
		"help",
		"h",
		false,
		"Print this help message",
	)

	var diffRef string
	pflag.StringVar(&diffRef, "ref", "", "an optional git ref to diff against")

	pflag.Parse()

	if helpRequested {
		printUsage()
		os.Exit(1)
	}

	checkData, err := checking.CheckChanges(diffRef)
	platform.FailOnErr(err)

	hasErrors := len(checkData.Errors) > 0
	hasWarnings := len(checkData.Warnings) > 0
	if hasErrors {
		fmt.Println("POTENTIAL MAJOR ISSUES:")
		for _, errorFlag := range checkData.Errors {
			fmt.Printf("  - %s\n", errorFlag.Message())
			if context := errorFlag.Context(); len(context) > 0 {
				fmt.Printf("    %s\n", context)
			}
		}
		if hasWarnings { fmt.Println("") }
	}

	if hasWarnings {
		fmt.Println("POTENTIAL ISSUES:")
		for _, warnFlag := range checkData.Warnings {
			fmt.Printf("  - %s\n", warnFlag.Message())
			if context := warnFlag.Context(); len(context) > 0 {
				fmt.Printf("    %s\n", context)
			}
		}
	}

	if hasErrors { os.Exit(1) }
}

// TODO: add information about all checks in here
func printUsage() {
	fmt.Fprint(
		os.Stderr,
		`Usage of check-changes:  check-changes [OPTION]...

Reads the current state of a git repository in the working directory, checking
for any potential things which you may want to know about before checking in
your code. Examples include added TODOs, mismatched indents, and more.

Flagged issues are divided into two levels of severity: major issues and
everything else.

Major issues are things which are almost always incorrect, and should prevent
code from being checked in at all. If a major issue is detected, check-changes
will exit with a non-zero status code. If used as a git hook, this will block
the commit from being created.

Non-major issues will be printed to standard output, but will exit with a zero
status code. Warnings are things which are valid to check in, but a programmer
may want to know about. For example: a TODO comment may be a good breadcrumb for
later work, but needs to be committed for now.

Options:
`,
	)
	pflag.PrintDefaults()
}
