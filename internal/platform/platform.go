package platform

import (
	"fmt"
	"os"
	"strings"
)

func FailOut(msg string) {
	fmt.Fprintln(os.Stderr, "ERROR: " + msg)
	os.Exit(1)
}

func FailOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: " + err.Error())
		os.Exit(1)
	}
}

func ErrMsg(msg string) string {
	return "ERROR: " + msg
}

func Assert(condition bool, more any) {
	if condition { return }
	panic(fmt.Sprintf("Assertion violated: %s", more))
}

func AssertNoErr(err error) {
	if err == nil { return }
	panic(fmt.Sprintf("Assertion violated, error encountered: %s ", err.Error()))
}

func SplitLines(s string) []string {
	return strings.FieldsFunc(s, func(c rune) bool {return c == '\n' || c == '\r'})
}
