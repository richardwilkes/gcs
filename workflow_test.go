// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

const (
	workflowDir         = ".github/workflows"
	releaseWorkflowPath = workflowDir + "/release.yml"
)

// workflowRunStep is one `run:` script extracted from a GitHub Actions workflow, along with the file and step name it
// came from so a failure can point at the offending step.
type workflowRunStep struct {
	file   string
	name   string
	script string
}

// TestWorkflowRunStepsDoNotInterpolateExpressions verifies that no workflow splices a `${{ ... }}` expression into a
// `run:` script. Actions substitutes those expressions before the shell ever sees the script, so a value containing
// quotes, `;` or `$(...)` becomes shell code executed by the job -- and the release job holds the macOS signing
// certificate, the notarization API key, and `contents: write`. Values belong in an `env:` entry and are referenced as
// "$VAR" from the script, which the shell treats as data no matter what it contains.
func TestWorkflowRunStepsDoNotInterpolateExpressions(t *testing.T) {
	c := check.New(t)
	entries, err := os.ReadDir(workflowDir)
	c.NoError(err)
	found := false
	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}
		found = true
		var steps []workflowRunStep
		steps, err = extractWorkflowRunSteps(filepath.Join(workflowDir, entry.Name()))
		c.NoError(err)
		for _, step := range steps {
			c.False(strings.Contains(step.script, "${{"),
				step.file+": step "+step.name+" interpolates an expression into its run script; pass it through env: instead")
		}
	}
	c.True(found, "no workflow files were scanned")
}

// TestReleaseWorkflowVersionStepResistsInjection runs the release workflow's "Determine version" script exactly as it
// is committed, feeding it a hostile version through the environment. The version has to survive as literal text and
// must not be able to run commands in the job.
func TestReleaseWorkflowVersionStepResistsInjection(t *testing.T) {
	c := check.New(t)
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	var steps []workflowRunStep
	steps, err = extractWorkflowRunSteps(releaseWorkflowPath)
	c.NoError(err)
	var script string
	for _, step := range steps {
		if step.name == "Determine version" {
			script = step.script
		}
	}
	c.NotEqual("", script, "the release workflow has no \"Determine version\" step")

	// The payloads below are only meaningful if they would in fact run when spliced into the script, so first confirm
	// that a version substituted directly into the shell -- what the step used to do -- executes the command it smuggles
	// in. Without this, the assertions that follow could pass simply because the payload was inert.
	dir := t.TempDir()
	_, err = runShellScript(bash, dir, `version="x"; touch pwned; echo ""`+"\n"+`echo "version=$version" >> "$GITHUB_OUTPUT"`, nil)
	c.NoError(err)
	_, err = os.Stat(filepath.Join(dir, "pwned"))
	c.NoError(err, "the injection payload is inert, so the checks below prove nothing")

	for _, tc := range []struct {
		name    string
		input   string
		ref     string
		refName string
		want    string
		wantErr bool
	}{
		{name: "dispatch input wins", input: "5.42.0", ref: "refs/heads/main", want: "5.42.0"},
		{name: "tag drives the version", ref: "refs/tags/v5.42.0", refName: "v5.42.0", want: "5.42.0"},
		// `build.sh -d` has no usable default (see TestBuildScriptRefusesDistWithoutRelease), so a run that can supply
		// no version has to stop here instead of handing an empty GCS_RELEASE to every build runner.
		{name: "no input, no tag", ref: "refs/heads/main", wantErr: true},
		{name: "quote escape", input: `x"; touch pwned; echo "`, ref: "refs/heads/main", want: `x"; touch pwned; echo "`},
		{name: "command substitution", input: "$(touch pwned)", ref: "refs/heads/main", want: "$(touch pwned)"},
		{name: "backtick substitution", input: "`touch pwned`", ref: "refs/heads/main", want: "`touch pwned`"},
		{name: "statement separator", input: "5.42.0; touch pwned", ref: "refs/heads/main", want: "5.42.0; touch pwned"},
	} {
		t.Run(tc.name, func(_ *testing.T) {
			runDir := t.TempDir()
			var version string
			version, err = runShellScript(bash, runDir, script, []string{
				"INPUT_VERSION=" + tc.input,
				"GITHUB_REF=" + tc.ref,
				"GITHUB_REF_NAME=" + tc.refName,
			})
			if tc.wantErr {
				c.HasError(err, tc.name)
			} else {
				c.NoError(err, tc.name)
				c.Equal(tc.want, version, tc.name)
			}
			_, err = os.Stat(filepath.Join(runDir, "pwned"))
			c.True(os.IsNotExist(err), tc.name+": the version was executed as shell code")
		})
	}
}

// TestBuildScriptRefusesDistWithoutRelease pins down the contract the release workflow's "Determine version" step
// depends on: `build.sh -d` has no built-in default version and refuses to run without GCS_RELEASE. The trailing -h is
// what keeps this cheap and safe -- argument processing is a single ordered pass, so -d either aborts first or, if the
// guard were ever dropped, -h prints the usage text and exits before anything is built.
func TestBuildScriptRefusesDistWithoutRelease(t *testing.T) {
	c := check.New(t)
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	cmd := exec.Command(bash, "build.sh", "-d", "-h")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
	out, err := cmd.CombinedOutput()
	c.HasError(err, "build.sh -d accepted an empty GCS_RELEASE")
	c.Contains(string(out), "GCS_RELEASE must be set")
}

// runShellScript runs a workflow step's script in dir with the given extra environment, then returns the value the
// script wrote to $GITHUB_OUTPUT as "version=...".
func runShellScript(bash, dir, script string, env []string) (string, error) {
	outputPath := filepath.Join(dir, "github_output")
	if err := os.WriteFile(outputPath, nil, 0o600); err != nil {
		return "", err
	}
	scriptPath := filepath.Join(dir, "step.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return "", err
	}
	// -e matches how Actions invokes a `shell: bash` step.
	cmd := exec.Command(bash, "-e", scriptPath)
	cmd.Dir = dir
	cmd.Env = append([]string{"PATH=" + os.Getenv("PATH"), "GITHUB_OUTPUT=" + outputPath}, env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", &shellError{output: string(out), err: err}
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return "", err
	}
	for line := range strings.Lines(strings.ReplaceAll(string(data), "\r\n", "\n")) {
		if after, ok := strings.CutPrefix(strings.TrimSuffix(line, "\n"), "version="); ok {
			return after, nil
		}
	}
	return "", nil
}

// shellError reports a failed script along with whatever it printed, since the exit status alone says nothing about
// what went wrong.
type shellError struct {
	err    error
	output string
}

func (e *shellError) Error() string {
	return e.err.Error() + ": " + e.output
}

func (e *shellError) Unwrap() error {
	return e.err
}

// extractWorkflowRunSteps pulls every `run:` script out of a GitHub Actions workflow file, remembering the name of the
// step each one came from. This is a deliberately small, indentation-based scan rather than a real YAML parse: the
// workflows use only the plain `run: <command>` and `run: |` block forms, and keeping it dependency-free avoids
// promoting a YAML package to a direct module requirement for the sake of one test.
func extractWorkflowRunSteps(path string) ([]workflowRunStep, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var steps []workflowRunStep
	var name string
	for i := 0; i < len(lines); {
		line := lines[i]
		i++
		key := strings.TrimPrefix(strings.TrimSpace(line), "- ")
		switch {
		case strings.HasPrefix(key, "name:"):
			name = strings.TrimSpace(strings.TrimPrefix(key, "name:"))
		case strings.HasPrefix(key, "run: |"):
			indent := indentOf(line)
			var block []string
			for i < len(lines) && (strings.TrimSpace(lines[i]) == "" || indentOf(lines[i]) > indent) {
				block = append(block, lines[i])
				i++
			}
			steps = append(steps, workflowRunStep{file: path, name: name, script: dedent(block)})
		case strings.HasPrefix(key, "run:") && strings.TrimSpace(strings.TrimPrefix(key, "run:")) != "":
			steps = append(steps, workflowRunStep{
				file:   path,
				name:   name,
				script: strings.TrimSpace(strings.TrimPrefix(key, "run:")),
			})
		}
	}
	return steps, nil
}

// indentOf returns the number of leading spaces on a line. Tabs are not valid YAML indentation, so spaces are all that
// need to be counted.
func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// dedent removes the common leading indentation from a block scalar's lines and joins them back together.
func dedent(block []string) string {
	indent := -1
	for _, line := range block {
		if strings.TrimSpace(line) != "" {
			if i := indentOf(line); indent == -1 || i < indent {
				indent = i
			}
		}
	}
	if indent < 1 {
		return strings.Join(block, "\n")
	}
	result := make([]string, 0, len(block))
	for _, line := range block {
		if len(line) < indent {
			result = append(result, strings.TrimLeft(line, " "))
		} else {
			result = append(result, line[indent:])
		}
	}
	return strings.Join(result, "\n")
}
