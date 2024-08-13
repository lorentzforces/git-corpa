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
	diffFiles, err := parseDiffLines(rawDiffLines)
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
var extendedHeaderRegex = regexp.MustCompile(`\A(index|mode|new file mode|deleted file mode)`)
func parseDiffLines(rawLines []string) ([]DiffFile, error) {
	files := make(map[string]DiffFile, 0)
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
				continue // impossible for there to be added lines here
			}

			// TODO: determine line number (in file) from diff

			lastHeaderHeader = i
			fileName = targetFile
			files[fileName] = DiffFile{
				FileName: fileName,
				Indents: IndentUnknown,
				ChangedLines: make([]DiffLine, 0),
			}
			continue
		}
		isExtendedHeader := extendedHeaderRegex.MatchString(rawLine)
		if isExtendedHeader {
			continue
		}
		// TODO: handle git's combined format that can have a variable number of "@" symbols
		isChunkHeader := strings.HasPrefix(rawLine, "@@")
		if isChunkHeader {
			lastChunkHeader = i
			continue
		}
		isHeaderLine := lastHeaderHeader < i && lastChunkHeader < lastHeaderHeader
		if isHeaderLine {
			continue
		}

		lineRunes := []rune(rawLine)
		firstChar := lineRunes[0]
		if firstChar == ' ' || firstChar == '-' {
			continue
		}
		platform.Assert(
			firstChar == '+',
			fmt.Sprintf(
				"Diff file content line did not start with one of [ -+], last file header: " +
					"\"%s\" offending line is:\n\"%s\"",
				fileName, rawLine,
			),
		)

		line := DiffLine{}
		CharLoop: for i := 1; i < len(lineRunes); i++ {
			ch := lineRunes[i]
			switch ch {
			case ' ': accreteIndentKind(line.Indents, IndentSpace)
			case '\t': accreteIndentKind(line.Indents, IndentTab)
			default: break CharLoop
			}
		}
		if lineRunes[1] == ' ' {
			line.Indents = IndentSpace
		} else if lineRunes[1] == '\t' {
			line.Indents = IndentTab
		}
	}

	fileSlice := make([]DiffFile, 0, len(files))
	for _, file := range files {
		fileSlice = append(fileSlice, file)
	}
	return fileSlice, nil
}

func accreteIndentKind(existing, observed IndentKind) IndentKind {
	if existing == IndentMixedLine {
		return IndentMixedLine
	}
	if existing == IndentUnknown {
		return observed
	}
	if observed == IndentUnknown {
		return existing
	}
	if existing == observed {
		return observed
	}

	return IndentMixedLine
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
	IndentUnknown IndentKind = iota
	IndentTab
	IndentSpace
	IndentMixedLine
)

func (ik IndentKind) String() string {
	switch ik {
		case IndentUnknown: return "IndentUnknown"
		case IndentTab: return "IndentTab"
		case IndentSpace: return "IndentSpace"
		case IndentMixedLine: return "IndentMixedLine"
	}
	platform.Assert(false, fmt.Sprintf("Invalid IndentKind value provided: %d", ik))
	panic("INVALID STATE: INVALID IndentKind VALUE PROVIDED")
}

type StashEntry struct {
	Number uint
	Branch string
	RawString string
}
