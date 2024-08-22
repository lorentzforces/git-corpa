package main

import (
	"fmt"
	"os"

	"github.com/lorentzforces/check-changes/internal/checking"
	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
)

// NOTE: for now, this only checks staged changes
// TODO: accept a ref and diff against that ref
// TODO: added todo line 1
// TODO: added todo line 2
func main() {
	if !git.ExecExists() {
		platform.FailOut("\"git\" executable not found on system PATH")
	}

	checkData, err := checking.CheckChanges()
	platform.FailOnErr(err)

	hasErrors := len(checkData.Errors) > 0
	if hasErrors {
		fmt.Println("POTENTIAL MAJOR ISSUES:")
		for _, errorFlag := range checkData.Errors {
			fmt.Printf("  - %s\n", errorFlag.Message())
			fmt.Println("")
		}
	}

	if len(checkData.Warnings) > 0 {
		fmt.Println("POTENTIAL ISSUES:")
		for _, warnFlag := range checkData.Warnings {
			fmt.Printf("  - %s\n", warnFlag.Message())
			fmt.Println("")
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}
