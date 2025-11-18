package gowok

import (
	"flag"

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
