package main

import (
	"fmt"

	"github.com/lorentzforces/check-changes/internal/checking"
	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
)

// bad news, based on what I'm seeing from libraries, we're going to have to shell out for everything and parse git output manually
// good news is that means we get to define all our own data structures and output

// NOTE: for now, this only checks staged changes
// TODO: accept a ref and diff against that ref
func main() {
	if !git.ExecExists() {
		platform.FailOut("\"git\" executable not found on system PATH")
	}

	checkData, err := checking.CheckChanges()
	platform.FailOnErr(err)
	fmt.Printf("checked data: %+v\n", checkData)




	// gather shared information
	// take shared information and iterate over the list of checks
	// checks:
	// - checks which operate on changed lines
	// - checks which operate on stash entries
	// checks return check result structure
	fmt.Printf("changes will be checked!\n")

}
