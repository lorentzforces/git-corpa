package checking

import (
	_ "embed"
	"errors"
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
	diffFiles, err := parseDiffLines(platform.SplitLines(pawtucketTest))
	if !assert.Nil(t, err) { t.FailNow() }
	if !assert.Len(t, diffFiles, 1) { t.FailNow() }
	for _, diffFile := range diffFiles {
		assert.Equal(t, IndentUnknown, diffFile.Indents,
			"diff file %s reported %s indents instead of IndentUnknown",
			diffFile.FileName, diffFile.Indents,
		)
	}
}

func TestParseDiffLines(t *testing.T) {
	diffFiles, err := parseDiffLines(platform.SplitLines(pawtucketTest))
	if !assert.Nil(t, err) { t.FailNow() }
	if !assert.Len(t, diffFiles, 1) { t.FailNow() }
	diffFile := diffFiles[0]

	assert.Equal(t, IndentUnknown, diffFile.Indents)
}
