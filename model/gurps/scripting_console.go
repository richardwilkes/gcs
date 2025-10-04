// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"fmt"
	"log/slog"
	"strings"
)

type scriptConsole struct{}

func (c scriptConsole) Log(args ...any) {
	slog.Info(makeLogMsg(args...))
}

func (c scriptConsole) Error(args ...any) {
	slog.Error(makeLogMsg(args...))
}

func makeLogMsg(args ...any) string {
	var buffer strings.Builder
	for argNum, arg := range args {
		if argNum > 0 {
			buffer.WriteByte(' ')
		}
		fmt.Fprintf(&buffer, "%v", arg)
	}
	return buffer.String()
}
