package src

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Encrypt struct {
	cfg      Config
	input    string
	password string
}

func (cmd Encrypt) Command() *cobra.Command {
	c := &cobra.Command{
		Use:     "encrypt",
		Aliases: []string{"e"},
		Short:   "Encrypt secrets",
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.run()
		},
	}
	c.Flags().StringVarP(&cmd.input, "input", "i", "envrypt.yaml", "path to the YAML file")
	c.Flags().StringVarP(&cmd.password, "password", "p", "", "password to use")
	must(c.MarkFlagRequired("password"))
	return c
}

func (cmd Encrypt) run() error {
	doc, err := Read(cmd.input)
	if err != nil {
		return fmt.Errorf("read %s: %v", cmd.input, err)
	}
	err = doc.Encrypt(cmd.password)
	if err != nil {
		return fmt.Errorf("encrypt %s: %v", cmd.input, err)
	}
	err = doc.Write()
	if err != nil {
		return fmt.Errorf("save %s: %v", cmd.input, err)
	}
	fmt.Fprintln(cmd.cfg.Stdout, "✅ encrypted")
	return nil
}
