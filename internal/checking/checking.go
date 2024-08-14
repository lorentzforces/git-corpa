package checking

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	diffFiles := parseDiffLines(rawDiffLines)
	platform.AssertNoErr(err)

	for i := range diffFiles {
		diffFile := &diffFiles[i]
		realFilePath := filepath.Join(repoRoot, diffFile.FileName)
		file, err := os.Open(realFilePath)
		platform.AssertNoErr(err)

		populateFileInfo(diffFile, file)
		file.Close()
	}
	checkData.Files = diffFiles

	// TODO: validate output check data

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
var chunkHeaderLineNumRegex = regexp.MustCompile(`\b+(\d+),`)
func parseDiffLines(rawLines []string) []DiffFile {
	files := make(map[string]*DiffFile, 0)
	fileName := ""
	lastHeaderHeader := -1
	lastChunkHeader := -1
	newFileLineNumber := -1

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

			lastHeaderHeader = i
			fileName = targetFile
			files[fileName] = &DiffFile{
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

		isChunkHeader := strings.HasPrefix(rawLine, "@@")
		if isChunkHeader {
			lastChunkHeader = i
			matches := chunkHeaderLineNumRegex.FindStringSubmatch(rawLine)
			if matches == nil {
				// chunk is removed lines only
				continue
			}
			lineNumber, err := strconv.Atoi(matches[1])
			platform.AssertNoErr(err)
			newFileLineNumber = lineNumber

			continue
		}

		isHeaderLine := lastHeaderHeader < i && lastChunkHeader < lastHeaderHeader
		if isHeaderLine {
			continue
		}

		lineRunes := []rune(rawLine)
		firstChar := lineRunes[0]
		if firstChar == ' ' {
			newFileLineNumber++
			continue
		}
		if firstChar == '-' {
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
		line.Content = rawLine
		line.LineNumber = uint(newFileLineNumber)
		line.Indents = whichLineIndents(lineRunes[1:])

		currentFile := files[fileName]
		currentFile.ChangedLines = append(currentFile.ChangedLines, line)
		newFileLineNumber++
	}

	fileSlice := make([]DiffFile, 0, len(files))
	for _, file := range files {
		fileSlice = append(fileSlice, *file)
	}

	return fileSlice
}

// do filesystem-related things with a DiffFile
func populateFileInfo(diffFile *DiffFile, file io.Reader) {
	input := bufio.NewScanner(file)

	fileIndents := IndentUnknown
	for input.Scan() && fileIndents == IndentUnknown {
		fileIndents = whichLineIndents([]rune(input.Text()))
	}
	diffFile.Indents = fileIndents
}

func whichLineIndents(line []rune) IndentKind {
	indents := IndentUnknown
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch ch {
		case ' ': indents = accreteIndentKind(indents, IndentSpace)
		case '\t': indents = accreteIndentKind(indents, IndentTab)
		default:
			return indents
		}
	}
	return indents
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
