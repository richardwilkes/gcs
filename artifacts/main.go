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
)

func main() {
    exe, err := os.Executable()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    dir := filepath.Dir(exe)
    var cmdArgs []string
    if runtime.GOOS == "darwin" {
        cmdArgs = append(cmdArgs, "-Xdock:icon="+filepath.Join(filepath.Dir(dir), "Resources", "app.png"))
    }
    cmdArgs = append(cmdArgs, "-m", "com.trollworks.gcs/com.trollworks.gcs.app.GCS")
    cmdArgs = append(cmdArgs, os.Args[1:]...)
    target := filepath.Join(dir, "support", "bin", "gcs")
    if runtime.GOOS == "windows" {
        target += ".exe"
    }
    cmd := exec.Command(target, cmdArgs...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err = cmd.Run(); err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            os.Exit(exitError.ExitCode())
        }
        fmt.Println(err)
        os.Exit(1)
    }
    os.Exit(0)
}
