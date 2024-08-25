package config

import (
	"os"

	"github.com/spf13/pflag"
)

// TODO: add some testing in here

type Config struct {
	HelpRequested bool
	HideContext bool
	DiffRef string
}

func Default() Config {
	return Config{}
}

func ApplyEnv(opts *Config) {
	opts.DiffRef = os.Getenv("CHK_CHNG_REF")
}

func InitOpts(opts *Config) *pflag.FlagSet {
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
		&opts.DiffRef,
		"ref",
		opts.DiffRef,
		"an optional git ref to diff against",
	)

	return flags
}

const EnvVarUsage = ``
