package src

import (
	"github.com/spf13/cobra"
)

type Encrypt struct {
	cfg Config
}

func (cmd Encrypt) Command() *cobra.Command {
	return &cobra.Command{
		Use:     "encrypt",
		Aliases: []string{"e"},
		Short:   "Encrypt secrets",
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.run()
		},
	}
}

func (cmd Encrypt) run() error {
	return nil
}
