package gurps

import (
	"fmt"
	"log/slog"
	"strings"
)

type scriptConsole struct{}

func (c scriptConsole) Log(args ...any) {
	var buffer strings.Builder
	for argNum, arg := range args {
		if argNum > 0 {
			buffer.WriteByte(' ')
		}
		fmt.Fprintf(&buffer, "%v", arg)
	}
	slog.Info(buffer.String())
}
