package main

import (
	"fmt"

	// "github.com/davecgh/go-spew/spew"
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

	for _, file := range checkData.Files {
		fmt.Printf("==DEBUG== diffed file %s has indents: %s\n", file.FileName, file.Indents)
	}
}
