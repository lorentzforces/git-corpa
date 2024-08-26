package config

import (
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

func ApplyEnv(opts *Opts) {
	opts.RawRevs = os.Getenv("CHK_CHNG_REF")
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

const rawRevsHelp string =
	"An optional git rev to diff against. You may pass a list of revs, separated by colons (:).\n" +
	"The first valid rev will be used.\n" +
	"If no valid rev is matched, the diff used will be as \"git diff\" with no arguments"

func (opts *Opts) ParseRevs() {
	opts.ParsedRevs = strings.Split(opts.RawRevs, ":")
}
