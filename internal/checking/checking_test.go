package checking

import (
	_ "embed"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/lorentzforces/check-changes/internal/platform"
	"github.com/stretchr/testify/assert"
)

func TestParsingStashEntries(t *testing.T) {
	cases := []struct{
		rawEntry string
		expectedNumber uint
		expectedBranch string
	} {
		{
			rawEntry: "stash@{0}: WIP on main: 12abc5 this doesn't matter",
			expectedNumber: 0,
			expectedBranch: "main",
		},
		{
			rawEntry: "stash@{1}: On ABC-1234-feature: 12abc5 this doesn't matter",
			expectedNumber: 1,
			expectedBranch: "ABC-1234-feature",
		},
		{
			rawEntry: "stash@{500}: On (no branch): 12abc5 this doesn't matter",
			expectedNumber: 500,
			expectedBranch: "(no branch)",
		},
	}

	for _, testCase := range cases {
		output, err := parseStashEntry(testCase.rawEntry)
		assert.Nil(t, err)

		if testCase.expectedBranch != output.Branch {
			t.Errorf(
				"parseStashEntry was expected to parse the branch \"%s\", but found the branch " +
					"\"%s\" instead",
				testCase.expectedBranch, output.Branch,
			)
		}

		if testCase.expectedNumber != output.Number {
			t.Errorf(
				"parseStashEntry was expected to parse the stash number %d, but found %d instead",
				testCase.expectedNumber, output.Number,
			)
		}
	}
}

// most of these are never happening in practice, but this helps me verify my parsing is working
// how I expect it to work
func TestParseStashEntriesWithErrors(t *testing.T) {
	rawEntries := []string {
		"an absolutely garbage stash entry",
		"stash@{-999}: On test: a negative number in the stash number",
		"stash@{nope}: On test: a non-number in the stash number",
	}

	for _, rawEntry := range rawEntries {
		_, err := parseStashEntry(rawEntry)
		if err == nil {
			t.Errorf(
				"Expected an error with this stash entry string, but was not given an error.\n" +
					"Offending entry: \"%s\"",
				rawEntry,
			)
		} else if errors.Is(err, stashParseError) {
			t.Logf("Expected error: %s", err.Error())
		} else {
			t.Errorf(
				"Expected an error matching 'stashParseError', but was given an unrelated error: %s",
				err.Error(),
			)
		}
	}
}

func TestAccreteIndentKind(t *testing.T) {
	cases := []struct{
		existing IndentKind
		observed IndentKind
		expected IndentKind
	} {
		{ IndentUnknown, IndentUnknown, IndentUnknown },
		{ IndentUnknown, IndentSpace, IndentSpace },
		{ IndentUnknown, IndentTab, IndentTab },
		{ IndentUnknown, IndentMixedLine, IndentMixedLine },
		{ IndentSpace, IndentUnknown, IndentSpace },
		{ IndentSpace, IndentSpace, IndentSpace },
		{ IndentSpace, IndentTab, IndentMixedLine },
		{ IndentSpace, IndentMixedLine, IndentMixedLine },
		{ IndentTab, IndentUnknown, IndentTab },
		{ IndentTab, IndentSpace, IndentMixedLine },
		{ IndentTab, IndentTab, IndentTab },
		{ IndentTab, IndentMixedLine, IndentMixedLine },
		{ IndentMixedLine, IndentUnknown, IndentMixedLine },
		{ IndentMixedLine, IndentSpace, IndentMixedLine },
		{ IndentMixedLine, IndentTab, IndentMixedLine },
		{ IndentMixedLine, IndentMixedLine, IndentMixedLine },
	}

	for _, testCase := range cases {
		result := accreteIndentKind(testCase.existing, testCase.observed)
		if testCase.expected != result {
			t.Errorf(
				"accreteIndentKind(%s, %s) produced %s, but %s was expected",
				testCase.existing, testCase.observed, result, testCase.expected,
			)
		}
	}
}

//go:embed pawtucket-test.diff
var pawtucketTest string

func TestParseDiffLinesAlwaysSetsUnknownIndents(t *testing.T) {
	diffFiles := parseDiffLines(platform.SplitLines(pawtucketTest))
	assert.Len(t, diffFiles, 1)
	if t.Failed() { t.FailNow() }

	for _, diffFile := range diffFiles {
		assert.Equal(t, IndentUnknown, diffFile.Indents,
			"diff file %s reported %s indents instead of IndentUnknown",
			diffFile.FileName, diffFile.Indents,
		)
	}
}

func TestParseDiffLines(t *testing.T) {
	diffFiles := parseDiffLines(platform.SplitLines(pawtucketTest))
	assert.Len(t, diffFiles, 1)
	if t.Failed() { t.FailNow() }

	diffFile := diffFiles[0]
	assert.Equal(t, "nantucket.txt", diffFile.FileName)

	assert.Len(t, diffFile.ChangedLines, 3)
	if t.Failed() { t.FailNow() }

	// first changed line
	firstChangeLine := diffFile.ChangedLines[0]
	assert.Equal(t, uint(1), firstChangeLine.LineNumber)
	assert.Equal(t, IndentUnknown, firstChangeLine.Indents)

	// second changed line
	secondChangeLine := diffFile.ChangedLines[1]
	assert.Equal(t, uint(3), secondChangeLine.LineNumber)
	assert.Equal(t, IndentTab, secondChangeLine.Indents)

	// third changed line
	thirdChangeLine := diffFile.ChangedLines[2]
	assert.Equal(t, uint(5), thirdChangeLine.LineNumber)
	assert.Equal(t, IndentUnknown, thirdChangeLine.Indents)
}

func TestPopulateFileInfo(t *testing.T) {
	spaceFile := strings.NewReader(`file header
    test
`)
	tabFile := strings.NewReader(`file header
	test
`)
	mixedFile := strings.NewReader(`file header
	  test
`)
	testCases := []struct {
		inputReader io.Reader
		expectedIndent IndentKind
	} {
		{ spaceFile, IndentSpace },
		{ tabFile, IndentTab },
		{ mixedFile, IndentMixedLine },
	}

	for _, testCase := range testCases {
		diffFile := &diffFile{}
		populateFileInfo(diffFile, testCase.inputReader)
		assert.Equal(t, testCase.expectedIndent, diffFile.Indents,
			"Expected indent kind %s, but observed %s",
			testCase.expectedIndent, diffFile.Indents,
		)
	}
}

func TestKeywordDetection(t *testing.T) {
	testData := checkData{
		Files: []diffFile{
			diffFile{
				FileName: "test-warn.txt",
				Indents: IndentSpace,
				ChangedLines: []diffLine{
					diffLine{
						LineNumber: 4,
						Indents: IndentSpace,
						Content: "// TODO: rewrite in rust",
					},
					diffLine{
						LineNumber: 5,
						Indents: IndentSpace,
						Content: "public static void main(string[] args)",
					},
				},
			},
			diffFile{
				FileName: "test-error.txt",
				Indents: IndentSpace,
				ChangedLines: []diffLine{
					diffLine{
						LineNumber: 6,
						Indents: IndentSpace,
						Content: "// NOCHECKIN",
					},
				},
			},
		},
	}

	result := reportChecks(testData)

	assert.Len(t, result.Warnings, 1, "expected exactly one warning-level flag")
	assert.Len(t, result.Errors, 1, "expected exactly one error-level flag")
	if t.Failed() { t.FailNow() }

	assert.IsType(t, KeywordPresenceFlag{}, result.Warnings[0])
	assert.IsType(t, KeywordPresenceFlag{}, result.Errors[0])
	if t.Failed() { t.FailNow() }
	warnFlag := result.Warnings[0].(KeywordPresenceFlag)
	errFlag := result.Errors[0].(KeywordPresenceFlag)

	assert.Equal(t, "test-warn.txt", warnFlag.FileName)
	assert.Equal(t, "TODO", warnFlag.Keyword)
	assert.Equal(t, uint(4), warnFlag.LineNumber)

	assert.Equal(t, "test-error.txt", errFlag.FileName)
	assert.Equal(t, "NOCHECKIN", errFlag.Keyword)
	assert.Equal(t, uint(6), errFlag.LineNumber)
}
