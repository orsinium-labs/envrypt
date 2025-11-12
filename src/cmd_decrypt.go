package src

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Decrypt struct {
	cfg      Config
	input    string
	output   string
	password string
}

func (cmd Decrypt) Command() *cobra.Command {
	c := &cobra.Command{
		Use:     "decrypt",
		Aliases: []string{"e"},
		Short:   "Decrypt secrets",
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.run()
		},
	}
	c.Flags().StringVarP(&cmd.input, "input", "i", "envrypt.yaml", "path to the YAML file")
	c.Flags().StringVarP(&cmd.output, "output", "o", ".env", "path to the YAML file")
	c.Flags().StringVarP(&cmd.password, "password", "p", "", "password to use")
	must(c.MarkFlagRequired("password"))
	return c
}

func (cmd Decrypt) run() error {
	doc, err := Read(cmd.input)
	if err != nil {
		return fmt.Errorf("read %s: %v", cmd.input, err)
	}
	vars, err := doc.Decrypt(cmd.password)
	if err != nil {
		return fmt.Errorf("encrypt %s: %v", cmd.input, err)
	}

	out, err := os.Create(cmd.output)
	if err != nil {
		return fmt.Errorf("open %s: %v", cmd.output, err)
	}
	defer func() {
		_ = out.Close()
	}()
	for _, envVar := range vars {
		val, _ := json.Marshal(envVar.Value)
		fmt.Fprintf(out, "export %s=%s\n", envVar.Key, string(val))
	}

	return nil
}
