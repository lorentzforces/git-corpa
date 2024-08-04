package checking

import (
	"errors"
	"testing"
)

func TestParsingStashEntries(T *testing.T) {
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

		if err != nil {
			T.Errorf("parseStashEntry returned unexpected error %s", err.Error())
		}

		if testCase.expectedBranch != output.Branch {
			T.Errorf(
				"parseStashEntry was expected to parse the branch \"%s\", but found the branch " +
					"\"%s\" instead",
				testCase.expectedBranch, output.Branch,
			)
		}

		if testCase.expectedNumber != output.Number {
			T.Errorf(
				"parseStashEntry was expected to parse the stash number %d, but found %d instead",
				testCase.expectedNumber, output.Number,
			)
		}
	}
}

// most of these are never happening in practice, but this helps me verify my parsing is working
// how I expect it to work
func TestParseStashEntriesWithErrors(T *testing.T) {
	rawEntries := []string {
		"an absolutely garbage stash entry",
		"stash@{-999}: On test: a negative number in the stash number",
		"stash@{nope}: On test: a non-number in the stash number",
	}

	for _, rawEntry := range rawEntries {
		_, err := parseStashEntry(rawEntry)
		if err == nil {
			T.Errorf(
				"Expected an error with this stash entry string, but was not given an error.\n" +
					"Offending entry: \"%s\"",
				rawEntry,
			)
		} else if errors.Is(err, stashParseError) {
			T.Logf("Expected error: %s", err.Error())
		} else {
			T.Errorf(
				"Expected an error matching 'stashParseError', but was given an unrelated error: %s",
				err.Error(),
			)
		}
	}
}
