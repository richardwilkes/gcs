package gurps

import (
	"fmt"
	"log/slog"
	"strings"
)

type scriptConsole struct{}

func (c *scriptConsole) Log(args ...any) {
	slog.Info(c.format(args...))
}

func (c *scriptConsole) Error(args ...any) {
	slog.Error(c.format(args...))
}

func (c *scriptConsole) Warn(args ...any) {
	slog.Warn(c.format(args...))
}

func (c *scriptConsole) Info(args ...any) {
	slog.Info(c.format(args...))
}

func (c *scriptConsole) Debug(args ...any) {
	slog.Debug(c.format(args...))
}

func (c *scriptConsole) format(args ...any) string {
	var buffer strings.Builder
	for argNum, arg := range args {
		if argNum > 0 {
			buffer.WriteByte(' ')
		}
		fmt.Fprintf(&buffer, "%v", arg)
	}
	return buffer.String()
}
