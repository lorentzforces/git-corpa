package main

import (
	"fmt"
	"os"

	"github.com/lorentzforces/check-changes/internal/checking"
	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
)

// NOTE: for now, this only checks staged changes
func main() {
	if !git.ExecExists() {
		platform.FailOut("\"git\" executable not found on system PATH")
	}

	checkData, err := checking.CheckChanges()
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

	if hasErrors {
		os.Exit(1)
	}
}
