package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/lorentzforces/check-changes/internal/platform"
)

func ExecExists() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

var RepoDoesNotExistError = fmt.Errorf("Not inside a git repository")
func RepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Env = []string{}

	stdOut, err := cmd.Output()
	exitErr, isType := err.(*exec.ExitError)

	if isType {
		stdErr := string(exitErr.Stderr[:])
		isNoRepoError := strings.Contains(stdErr, "not a git repository")
		if isNoRepoError {
			return "", RepoDoesNotExistError
		}
		platform.FailOnErr(err)
	}
	platform.FailOnErr(err)

	fullOutput := string(stdOut[:])
	return strings.TrimRight(fullOutput, "\n\r"), nil
}

// The current branch name. If in detached head state, returns empty string.
func CurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	stdOut, err := cmd.Output()
	platform.FailOnErr(err)
	fullOutput := string(stdOut[:])
	return strings.TrimRight(fullOutput, "\n\r")
}

func StashEntries() []string {
	cmd := exec.Command("git", "stash", "list")
	cmd.Env = []string{}
	stdOut, err := cmd.Output()
	platform.FailOnErr(err)

	fullOutput := string(stdOut[:])
	return platform.SplitLines(fullOutput)
}

// If ref is a non-empty string, diff against whatever ref that is.
// If empty, then just diff against HEAD (only index/staged changes)
func Diff(ref string) []string {
	refForDiff := ref
	if len(refForDiff) == 0 { refForDiff = "HEAD" }

	cmd := exec.Command("git", "diff", "--no-color", "-p", refForDiff)
	cmd.Env = []string{}
	stdOut, err := cmd.Output()
	platform.FailOnErr(err)

	fullOutput := string(stdOut[:])
	return platform.SplitLines(fullOutput)
}
