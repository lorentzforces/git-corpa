package checking

import "testing"

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

func TestParseStashEntriesWithErrors(T *testing.T) {
}
