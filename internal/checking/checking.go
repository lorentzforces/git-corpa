package checking

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
)

func CheckChanges() CheckData {
	checkData := CheckData{}
	checkData.CurrentBranch = git.CurrentBranch()

	rawStashList := git.StashEntries()
	fmt.Printf("raw stash entries: %+v", rawStashList)


	// TODO: get stash items
	// TODO: get the diff



	return checkData
}

func parseStashEntries() ([]StashEntry, error) {
	return nil, nil
}

// format: "stash@{N}: [WIP on|On] branchName:" followed by stuff we don't care about
// CAPTURE GROUPS (submatches)
// number: 1
// branch: 3
const stashEntryPattern string = "^stash@\\{(?<number>\\d+)}:( WIP)? [Oo]n (?<branch>[^:]+):"

func parseStashEntry(rawEntry string) (StashEntry, error) {
	stashRegex := regexp.MustCompile(stashEntryPattern)
	matches := stashRegex.FindStringSubmatch(rawEntry)
	if matches == nil {
		err := fmt.Errorf("Malformed input: output of 'git stash list' was unparseable")
		return StashEntry{}, err
	}

	rawNumber := matches[1]
	number, err := strconv.ParseUint(rawNumber, 10, 64)
	platform.AssertNoErr(err)
	return StashEntry{
			Number: uint(number),
			Branch: matches[3],
		},
		nil
}

// TODO: all these types are speculative

type CheckData struct {
	CurrentBranch string
	Files []DiffFile
	StashEntries []StashEntry
}

type DiffFile struct {
	FileName string
	Indents IndentKind
	ChangedLines []DiffLine
}

type DiffLine struct {
	LineNumber uint
	Indents IndentKind
	Content string
}

type IndentKind int
const (
	IndentTab IndentKind = iota
	IndentSpace
)

type StashEntry struct {
	Number uint
	Branch string
}
