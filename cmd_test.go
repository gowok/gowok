package gowok

import (
	"flag"
	"testing"

	"github.com/golang-must/must"
	"github.com/spf13/cobra"
)

func TestCmd_Flags(t *testing.T) {
	t.Run("positive/Flags singleton", func(t *testing.T) {
		f1 := Flags()
		f2 := Flags()
		must.NotNil(t, f1)
		must.Equal(t, f1, f2)
	})
}

func TestCmd_Wrap(t *testing.T) {
	t.Run("positive/Wrap adds flags", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		wrapped := CMD.Wrap(cmd)

		must.Equal(t, cmd, wrapped)

		must.NotNil(t, flag.Lookup("config"))
		must.NotNil(t, flag.Lookup("env-file"))
	})
}

func TestCmd_flagParse(t *testing.T) {
	t.Run("positive/flagParse idempotent", func(t *testing.T) {
		flagParse()
		flagParse()

		must.NotNil(t, flag.Lookup("config"))
	})
}
