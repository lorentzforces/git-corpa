package checking

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
)

func CheckChanges() CheckData {
	checkData := CheckData{}
	checkData.CurrentBranch = git.CurrentBranch()

	stashEntries, err := parseStashEntries(git.StashEntries())
	platform.AssertNoErr(err)
	checkData.StashEntries = stashEntries

	fmt.Printf("parsed stash entries: %+v\n", stashEntries)

	// TODO: get stash items
	// TODO: get the diff



	return checkData
}

func parseStashEntries(rawEntries []string) ([]StashEntry, error) {
	entries := make([]StashEntry, 0, len(rawEntries))
	for _, rawEntry := range rawEntries {
		entry, err := parseStashEntry(rawEntry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// format: "stash@{N}: [WIP on|On] branchName:" followed by stuff we don't care about
// CAPTURE GROUPS (submatches)
// number: 1
// branch: 3
const stashEntryPattern string = "^stash@\\{(?<number>\\d+)}:( WIP)? [Oo]n (?<branch>[^:]+):"
var stashParseError error = fmt.Errorf("An error was encountered while parsing a stash entry string")

func parseStashEntry(rawEntry string) (StashEntry, error) {
	stashRegex := regexp.MustCompile(stashEntryPattern)
	matches := stashRegex.FindStringSubmatch(rawEntry)
	if matches == nil {
		err := fmt.Errorf(
			"Malformed input: output of 'git stash list' was unparseable: \"%s\"",
			rawEntry,
		)
		return StashEntry{}, errors.Join(stashParseError, err)
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
