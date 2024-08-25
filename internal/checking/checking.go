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

func CheckChanges(diffRef string) (CheckReport, error) {
	checkData, err := gatherState(diffRef)
	if err != nil {
		return CheckReport{}, err
	}

	return reportChecks(checkData), nil
}

type CheckReport struct {
	Errors []CheckFlag
	Warnings []CheckFlag
}

type CheckFlag interface {
	Message() string
	Context() string
}

type StashEntryFlag struct {
	Number uint
	FullLine string
}

func (flag StashEntryFlag) Message() string {
	return fmt.Sprintf("Stash entry {%d} has stashed changes from your current branch", flag.Number)
}

func (flag StashEntryFlag) Context() string {
	return flag.FullLine
}

type LineIndentFlag struct {
	FileName string
	LineNumber uint
	FileIndents IndentKind
	LineIndents IndentKind
}

func (flag LineIndentFlag) Message() string {
	return fmt.Sprintf(
		"%s:%d | line has indentation (%s) inconsistent with the rest of the file (%s)",
		flag.FileName, flag.LineNumber, flag.FileIndents, flag.LineIndents,
	)
}

func (flag LineIndentFlag) Context() string {
	return ""
}

type KeywordPresenceFlag struct {
	FileName string
	LineNumber uint
	Keyword string
	LineContent string
}

func (flag KeywordPresenceFlag) Message() string {
	return fmt.Sprintf(
		"%s:%d | line contains keyword \"%s\"",
		flag.FileName, flag.LineNumber, flag.Keyword,
	)
}

func (flag KeywordPresenceFlag) Context() string {
	return trimReportedLine(flag.LineContent)
}

type checkData struct {
	CurrentBranch string
	Files []diffFile
	StashEntries []stashEntry
}

type diffFile struct {
	FileName string
	Indents IndentKind
	ChangedLines []diffLine
}

type diffLine struct {
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

func (ik IndentKind) incompatibleWithLine(lineIndent IndentKind) bool {
	if ik == IndentUnknown || lineIndent == IndentUnknown {
		return false
	}

	return ik != lineIndent
}

type stashEntry struct {
	Number uint
	Branch string
	RawString string
}

func gatherState(diffRef string) (checkData, error) {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return checkData{}, err
	}

	checkData := checkData{}
	checkData.CurrentBranch = git.CurrentBranch()

	stashEntries, err := parseStashEntries(git.StashEntries())
	platform.FailOnErr(err)
	checkData.StashEntries = stashEntries

	rawDiffLines := git.Diff(diffRef)

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

	return checkData, nil
}

func parseStashEntries(rawEntries []string) ([]stashEntry, error) {
	entries := make([]stashEntry, 0, len(rawEntries))
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
func parseStashEntry(rawEntry string) (stashEntry, error) {
	matches := stashEntryRegex.FindStringSubmatch(rawEntry)
	if matches == nil {
		err := fmt.Errorf(
			"Malformed input: output of 'git stash list' was unparseable: \"%s\"",
			rawEntry,
		)
		return stashEntry{}, errors.Join(stashParseError, err)
	}

	rawNumber := matches[1]
	number, err := strconv.ParseUint(rawNumber, 10, 64)
	platform.AssertNoErr(err)
	return stashEntry{
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

func parseDiffLines(rawLines []string) []diffFile {
	files := make(map[string]*diffFile, 0)
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
			files[fileName] = &diffFile{
				FileName: fileName,
				Indents: IndentUnknown,
				ChangedLines: make([]diffLine, 0),
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

		line := diffLine{}
		line.Content = rawLine
		line.LineNumber = uint(newFileLineNumber)
		line.Indents = whichLineIndents(lineRunes[1:])

		currentFile := files[fileName]
		currentFile.ChangedLines = append(currentFile.ChangedLines, line)
		newFileLineNumber++
	}

	fileSlice := make([]diffFile, 0, len(files))
	for _, file := range files {
		fileSlice = append(fileSlice, *file)
	}

	return fileSlice
}

// do filesystem-related things with a diffFile
func populateFileInfo(diffFile *diffFile, file io.Reader) {
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

var errorKeywords = map[string]struct{} {
	"NOCHECKIN": struct{}{},
}
var warnKeywords = map[string]struct{} {
	"TODO": struct{}{},
}

func initKeywordRegex(wordSets... map[string]struct{}) *regexp.Regexp {
	var buf strings.Builder
	_, _ = buf.WriteString(`\b(`)

	isFirst := true
	for _, words := range wordSets {
		for word, _ := range words {
			if !isFirst { _, _ = buf.WriteString(`|`) }
			isFirst = false
			_, _ = buf.WriteString(word)
		}
	}

	_, _ = buf.WriteString(`)\b`)
	return regexp.MustCompile(buf.String())
}

var keywordRegex = initKeywordRegex(warnKeywords, errorKeywords)

func reportChecks(data checkData) CheckReport {
	result := CheckReport{
		Errors: make([]CheckFlag, 0),
		Warnings: make([]CheckFlag, 0),
	}

	for _, entry := range data.StashEntries {
		if entry.Branch == data.CurrentBranch {
			result.Warnings = append(
				result.Warnings,
				StashEntryFlag{
					Number: entry.Number,
					FullLine: entry.RawString,
				},
			)
		}
	}

	for _, file := range data.Files {
		for _, line := range file.ChangedLines {
			if file.Indents.incompatibleWithLine(line.Indents) {
				result.Errors = append(
					result.Errors,
					LineIndentFlag {
						FileName: file.FileName,
						LineNumber: line.LineNumber,
						FileIndents: file.Indents,
						LineIndents: line.Indents,
					},
				)
			}

			keyword := keywordRegex.FindString(line.Content)
			keywordFlag := KeywordPresenceFlag{
				FileName: file.FileName,
				LineNumber: line.LineNumber,
				Keyword: keyword,
				LineContent: line.Content,
			}
			if _, ok := errorKeywords[keyword]; ok {
				result.Errors = append(result.Errors, keywordFlag)
			}
			if _, ok := warnKeywords[keyword]; ok {
				result.Warnings = append(result.Warnings, keywordFlag)
			}
		}
	}

	return result
}

// Remove git diff marker, any leading whitespace after that marker, and any trailing whitespace.
// Additionally, if the trimmed result is more than 80 characters, chop it down to 80
func trimReportedLine(line string) string {
	lineRunes := []rune(line)
	lineRunes = lineRunes[1:]
	trimmedLine := []rune(strings.TrimSpace(string(lineRunes)))

	var finalLine string
	if len(trimmedLine) > 80 {
		trimmedLine = trimmedLine[:77]
		finalLine = fmt.Sprintf("%s...", string(trimmedLine))
	} else {
		finalLine = string(trimmedLine)
	}

	return finalLine
}
