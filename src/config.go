package src

import (
	"io"
	"os"
)

// Config holds global configurations for commands
type Config struct {
	Stdin  io.Reader
	Stdout io.Writer
}

// Read implements io.Reader. It's a shortcut for reading from stdin.
func (c Config) Read(p []byte) error {
	_, err := c.Stdin.Read(p)
	return err
}

// Write implements io.Writer. It's a shortcut for writing into stdout.
func (c Config) Write(p []byte) error {
	_, err := c.Stdout.Write(p)
	return err
}

// HasStdin checks if there is data passed into stdin
func (c Config) HasStdin() bool {
	stdin, ok := c.Stdin.(*os.File)
	if !ok {
		return true
	}
	stat, _ := stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
