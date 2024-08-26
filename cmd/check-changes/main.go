package main

import (
	"fmt"
	"os"

	"github.com/lorentzforces/check-changes/internal/checking"
	"github.com/lorentzforces/check-changes/internal/config"
	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
)

func main() {
	opts := config.Default()
	config.ApplyEnv(&opts)
	flags := config.InitOpts(&opts)
	flags.Parse(os.Args[1:])

	if opts.HelpRequested {
		printUsage()
		os.Exit(1)
	}

	if !git.ExecExists() {
		platform.FailOut("\"git\" executable not found on system PATH")
	}

	checkData, err := checking.CheckChanges(opts.DiffRef)
	platform.FailOnErr(err)

	printResults(checkData)

	if len(checkData.Errors) > 0 { os.Exit(1) }
}

func printResults(results checking.CheckReport) {
	hasErrors := len(results.Errors) > 0
	hasWarnings := len(results.Warnings) > 0
	if hasErrors {
		fmt.Println("POTENTIAL MAJOR ISSUES:")
		for _, errorFlag := range results.Errors {
			fmt.Printf("  - %s\n", errorFlag.Message())
			if context := errorFlag.Context(); len(context) > 0 {
				fmt.Printf("    %s\n", context)
			}
		}
		if hasWarnings { fmt.Println("") }
	}

	if hasWarnings {
		fmt.Println("POTENTIAL ISSUES:")
		for _, warnFlag := range results.Warnings {
			fmt.Printf("  - %s\n", warnFlag.Message())
			if context := warnFlag.Context(); len(context) > 0 {
				fmt.Printf("    %s\n", context)
			}
		}
	}
}

// TODO: add information about all checks in here
// TODO: add information about environment variables
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

	flags := config.InitOpts(&config.Config{})
	flags.PrintDefaults()
}
