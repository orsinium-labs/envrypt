package src

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

func command(cfg Config) *cobra.Command {
	root := &cobra.Command{
		Use:   "envrypt",
		Short: "Encrypt environment variables.",
		Long: `
			CLI tool for storing encrypted production secrets in a file.
		`,
		SilenceUsage: true, // do not print usage when a command returns an error
	}
	root.AddCommand(Version{cfg: cfg}.Command())
	root.AddCommand(Encrypt{cfg: cfg}.Command())
	return root
}

func Main(args []string, r io.Reader, w io.Writer) error {
	cfg := Config{Stdin: r, Stdout: w}
	c := command(cfg)
	c.SetArgs(args)
	return c.Execute()
}

func Entrypoint() {
	err := Main(os.Args[1:], os.Stdin, os.Stdout)
	if err != nil {
		os.Exit(1)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
