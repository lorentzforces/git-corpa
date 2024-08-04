package git

import (
	"os/exec"
	"strings"

	"github.com/lorentzforces/check-changes/internal/platform"
)

func ExecExists() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func CurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	stdOut, err := cmd.Output()
	platform.AssertNoErr(err)

	// TODO: if in detached head, this will be empty
	// TODO: unsure if this output includes trailing newline? trim if it does
	return string(stdOut[:])
}

func StashEntries() []string {
	cmd := exec.Command("git", "stash", "list")
	stdOut, err := cmd.Output()
	platform.AssertNoErr(err)

	fullOutput := string(stdOut[:])

	// TODO: unsure if this output includes trailing newline? trim if it does
	return strings.FieldsFunc(fullOutput, func(c rune) bool {return c == '\n' || c == '\r'})
}
