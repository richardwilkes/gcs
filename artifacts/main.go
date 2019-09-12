/*
 * Copyright (c) 2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dir := filepath.Dir(exe)
	exeName := filepath.Base(exe)
	createLinuxDesktopFile(dir)
	os.Exit(launchApp(dir, exeName))
}

func createLinuxDesktopFile(dir string) {
	if runtime.GOOS == "linux" {
		if home := os.Getenv("HOME"); home != "" {
			desktopPath := filepath.Join(home, ".local", "share", "applications")
			if err := os.MkdirAll(desktopPath, 0775); err != nil {
				fmt.Println(err)
				return
			}
			desktopPath = filepath.Join(desktopPath, "gcs.desktop")
			f, err := os.Create(desktopPath)
			if err != nil {
				fmt.Println(err)
				return
			}
			if _, err = fmt.Fprintf(f, `[Desktop Entry]
Version=1.0
Type=Application
Name=GCS
Icon=%[1]s/support/app.png
Path=%[1]s
Exec=%[1]s/gcs %%F
Categories=Games;Roleplaying
Keywords=GURPS;GCS;Character;Sheet
Terminal=false
`, strings.ReplaceAll(dir, " ", `\ `)); err != nil {
				fmt.Println(err)
			}
			if cerr := f.Close(); cerr != nil && err == nil {
				fmt.Println(cerr)
				return
			}
			if err = os.Chmod(desktopPath, 0775); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func createCmdArgs(dir string) []string {
	var cmdArgs []string
	if runtime.GOOS == "darwin" {
		cmdArgs = append(cmdArgs, "-Xdock:icon="+filepath.Join(filepath.Dir(dir), "Resources", "app.png"))
	}
	cmdArgs = append(cmdArgs, "-m", "com.trollworks.gcs/com.trollworks.gcs.app.GCS")
	return append(cmdArgs, os.Args[1:]...)
}

func launchApp(dir, exeName string) int {
	cmd := exec.Command(filepath.Join(dir, "support", "bin", exeName), createCmdArgs(dir)...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		fmt.Println(err)
		return 1
	}
	return 0
}
