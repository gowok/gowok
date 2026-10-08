package gowok

import (
	"flag"

	"github.com/gowok/gowok/singleton"
	"github.com/spf13/cobra"
)

type _cmd struct {
	*cobra.Command
}

var CMD = &_cmd{
	Command: &cobra.Command{},
}

func (p *_cmd) Wrap(c *cobra.Command) *cobra.Command {
	flagParse()
	c.Flags().AddGoFlagSet(flag.CommandLine)
	return c
}

type flags struct {
	Config  string
	EnvFile string
	Help    bool
}

var _flags = singleton.New(func() *flags {
	return &flags{}
})

func Flags() *flags {
	return *_flags()
}

func flagParse() {
	if flag.Lookup("config") == nil {
		flag.StringVar(&Flags().Config, "config", "", "configuration file location (yaml, toml)")
	}
	if flag.Lookup("env-file") == nil {
		flag.StringVar(&Flags().EnvFile, "env-file", "", "env file location")
	}
}
