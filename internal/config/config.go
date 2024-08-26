package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// TODO: add some testing in here

type Opts struct {
	HelpRequested bool
	HideContext bool
	RawRevs string
	ParsedRevs []string
}

func Default() Opts {
	return Opts{}
}

func InitOpts(opts *Opts) *pflag.FlagSet {
	flags := pflag.NewFlagSet("main", pflag.ExitOnError)

	flags.BoolVarP(
		&opts.HelpRequested,
		"help",
		"h",
		opts.HelpRequested,
		"Print this help message",
	)
	flags.BoolVar(
		&opts.HideContext,
		"no-context",
		opts.HideContext,
		"Do not print additional context information with flagged issues",
	)
	flags.StringVar(
		&opts.RawRevs,
		"revs",
		opts.RawRevs,
		rawRevsHelp,
	)

	return flags
}

const envPrefix string = "CHCK_CHNG_"
const rawRevsEnv string = envPrefix + "REVS"

func ApplyEnv(opts *Opts) {
	opts.RawRevs = os.Getenv(rawRevsEnv)
}

const rawRevsHelp string =
	`An optional git rev to diff against. You may pass a list of revs, separated by colons (:).
	The first valid rev will be used.
	If no valid rev is matched, the diff used will be as \"git diff\" with no arguments.`

func (opts *Opts) ParseRevs() {
	opts.ParsedRevs = strings.Split(opts.RawRevs, ":")
}

func EnvVarHelp() string {
	text := `ENVIRONMENT VARIABLES

Some values can be set in environment variables. In general, an option which
can be specified by an environment variable will be overridden if that option
is specified as a command-line option.

The following environment variables are available:
  - %s: corresponds to the "revs" command-line option`

	return fmt.Sprintf(text, rawRevsEnv)
}
