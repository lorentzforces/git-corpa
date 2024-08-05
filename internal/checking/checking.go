package checking

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lorentzforces/check-changes/internal/git"
	"github.com/lorentzforces/check-changes/internal/platform"
)

func CheckChanges() (CheckData, error) {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return CheckData{}, err
	}

	checkData := CheckData{}
	checkData.CurrentBranch = git.CurrentBranch()

	stashEntries, err := parseStashEntries(git.StashEntries())
	platform.FailOnErr(err)
	checkData.StashEntries = stashEntries

	rawDiffLines := git.Diff()
	diffFiles, err := parseDiff(rawDiffLines)
	_, _ = diffFiles, err

	_ = repoRoot
	return checkData, nil
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

var stashParseError = fmt.Errorf("An error was encountered while parsing a stash entry string")

// format: "stash@{N}: [WIP on|On] branchName:" followed by stuff we don't care about
// CAPTURE GROUPS (submatches) number: 1, branch: 3
var stashEntryRegex = regexp.MustCompile(`^stash@\{(?<number>\d+)}:( WIP)? [Oo]n (?<branch>[^:]+):`)
func parseStashEntry(rawEntry string) (StashEntry, error) {
	matches := stashEntryRegex.FindStringSubmatch(rawEntry)
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
			RawString: rawEntry,
		},
		nil
}

var diffParseError = fmt.Errorf("An error was encoutnered while parsing diff output")

var resultFileRegex = regexp.MustCompile(`b/(\S+)\z`)
func parseDiff(rawLines []string) ([]DiffFile, error) {

	// header-header: starts with "diff --git"
	// header lines: between a header-header and the first chunk
	// chunk lines: start with @@





	files := make([]DiffFile, 0)
	fileName := ""
	_ = fileName
	lastHeaderHeader := 0
	lastChunkHeader := 0

	for i, rawLine := range rawLines {
		isHeaderHeader := strings.HasPrefix(rawLine, "diff --git")
		if isHeaderHeader {
			matches := resultFileRegex.FindStringSubmatch(rawLine)
			platform.Assert(
				matches != nil,
				fmt.Sprintf("Result file regex did not find a target file name on line %d", i),
			)

			targetFile := matches[1]
			if targetFile == "/dev/null" {
				// impossible for there to be added lines here
				continue
			}

			lastHeaderHeader = i
			fileName = targetFile
		}
		isChunkHeader := strings.HasPrefix(rawLine, "@@")
		if isChunkHeader {
			lastChunkHeader = i
			continue
		}
		isHeaderLine := lastHeaderHeader < i && lastChunkHeader < lastHeaderHeader
		if isHeaderLine {
			continue
		}

		firstChar := fmt.Sprintf("%.1s", rawLine)
		if firstChar == " " || firstChar == "-" {
			continue
		}
		platform.Assert(
			firstChar == "+",
			fmt.Sprintf(
				"Diff file content line did not start with one of [ -+], last file header: " +
					"\"%s\" offending line is:\n%s",
				fileName, rawLine,
			),
		)

		// TODO: add file struct where necessary
		// TODO: add line struct to files

		// TODO: figure out where original file whitespace determination happens? putting it in here makes testing a pain

	}
	return files, nil
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
	RawString string
}
