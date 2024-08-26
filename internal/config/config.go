package config

import (
	"os"

	"github.com/spf13/pflag"
)

// TODO: add some testing in here

type Opts struct {
	HelpRequested bool
	HideContext bool
	DiffRev string
}

func Default() Opts {
	return Opts{}
}

func ApplyEnv(opts *Opts) {
	opts.DiffRev = os.Getenv("CHK_CHNG_REF")
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
		&opts.DiffRev,
		"rev",
		opts.DiffRev,
		"an optional git ref to diff against",
	)

	return flags
}

const EnvVarUsage = ``
