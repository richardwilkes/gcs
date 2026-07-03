// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/go-git/go-billy/v6/memfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestDownloadLatestCommitReservedName verifies that downloadLatestCommit can check out a library whose files have base
// names that collide with Windows reserved device names (e.g. "Con Man.gct", whose "Con" prefix matches "CON"). go-git
// enables its NTFS/HFS+ path protections by default on every platform, which would otherwise reject the checkout with
// `openfile: invalid path: "Library/Space/Con Man.gct"` on macOS and Linux as well as Windows. See issue #1057.
func TestDownloadLatestCommitReservedName(t *testing.T) {
	c := check.New(t)

	const reservedPath = "Library/Space/Con Man.gct"
	const content = "hello con man\n"
	srcDir := t.TempDir()
	wantHash := buildBareLibraryRepo(t, srcDir, reservedPath, content)

	fs := memfs.New()
	hash, err := downloadLatestCommit(context.Background(), srcDir, "", fs)
	c.NoError(err, "downloadLatestCommit should tolerate Windows reserved device names")
	c.Equal(wantHash.String(), hash, "returned hash should match the source commit")

	f, err := fs.Open(reservedPath)
	c.NoError(err, "the reserved-name file should be checked out")
	if err == nil {
		data, readErr := io.ReadAll(f)
		c.NoError(readErr, "should be able to read the checked-out file")
		c.NoError(f.Close(), "should be able to close the checked-out file")
		c.Equal(content, string(data), "the checked-out file should have the expected content")
	}
}

// buildBareLibraryRepo creates a bare git repository at dir containing a single commit with one file at filePath holding
// content, and returns the commit hash. It writes the git objects directly rather than going through a worktree,
// because go-git's worktree operations apply the same reserved-name path validation that this test needs to place a
// file past.
func buildBareLibraryRepo(t *testing.T, dir, filePath, content string) plumbing.Hash {
	t.Helper()
	repo, err := git.PlainInit(dir, true)
	if err != nil {
		t.Fatalf("unable to init source repo: %v", err)
	}
	s := repo.Storer

	blob := s.NewEncodedObject()
	blob.SetType(plumbing.BlobObject)
	w, err := blob.Writer()
	if err != nil {
		t.Fatalf("unable to open blob writer: %v", err)
	}
	if _, err = io.WriteString(w, content); err != nil {
		t.Fatalf("unable to write blob: %v", err)
	}
	if err = w.Close(); err != nil {
		t.Fatalf("unable to close blob writer: %v", err)
	}
	hash := storeObject(t, s, blob)

	// Build the nested trees from the leaf up. Each path component becomes a directory tree except the last, which is
	// the file entry.
	parts := splitPath(filePath)
	name := parts[len(parts)-1]
	entry := object.TreeEntry{Name: name, Mode: filemode.Regular, Hash: hash}
	for i := len(parts) - 2; i >= 0; i-- {
		tree := &object.Tree{Entries: []object.TreeEntry{entry}}
		treeObj := s.NewEncodedObject()
		if err = tree.Encode(treeObj); err != nil {
			t.Fatalf("unable to encode tree: %v", err)
		}
		entry = object.TreeEntry{Name: parts[i], Mode: filemode.Dir, Hash: storeObject(t, s, treeObj)}
	}
	rootTree := &object.Tree{Entries: []object.TreeEntry{entry}}
	rootObj := s.NewEncodedObject()
	if err = rootTree.Encode(rootObj); err != nil {
		t.Fatalf("unable to encode root tree: %v", err)
	}
	rootHash := storeObject(t, s, rootObj)

	sig := object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()}
	commit := &object.Commit{Author: sig, Committer: sig, Message: "initial", TreeHash: rootHash}
	commitObj := s.NewEncodedObject()
	if err = commit.Encode(commitObj); err != nil {
		t.Fatalf("unable to encode commit: %v", err)
	}
	commitHash := storeObject(t, s, commitObj)

	if err = s.SetReference(plumbing.NewHashReference(plumbing.Master, commitHash)); err != nil {
		t.Fatalf("unable to set master reference: %v", err)
	}
	if err = s.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.Master)); err != nil {
		t.Fatalf("unable to set HEAD reference: %v", err)
	}
	return commitHash
}

func storeObject(t *testing.T, s storage.Storer, obj plumbing.EncodedObject) plumbing.Hash {
	t.Helper()
	hash, err := s.SetEncodedObject(obj)
	if err != nil {
		t.Fatalf("unable to store object: %v", err)
	}
	return hash
}

func splitPath(p string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(p); i++ {
		if p[i] == '/' {
			parts = append(parts, p[start:i])
			start = i + 1
		}
	}
	return append(parts, p[start:])
}
