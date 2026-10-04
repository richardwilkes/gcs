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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xreflect"
)

// LibraryFile holds the library and path to a file.
type LibraryFile struct {
	Library string `json:"library"`
	Path    string `json:"path"`
}

// SourcedID holds a TID and an optional Source.
type SourcedID struct {
	TID    tid.TID `json:"id"`
	Source Source  `json:"source,omitzero"`
}

// Source holds a reference to the source of a particular piece of data.
type Source struct {
	LibraryFile
	TID tid.TID `json:"id"`
}

type libSrcData struct {
	// path, timestamp and size identify the version of the file the hashes were loaded from.
	path       string
	timestamp  time.Time
	size       int64
	dataHashes map[tid.TID]HashAndData
}

// SrcProvider defines the methods needed for a source provider that can be used with the SrcMatcher.Match() function.
type SrcProvider interface {
	Hashable
	GetSource() Source
}

// SrcMatcher provides Source matching for a given ListProvider.
type SrcMatcher struct {
	libHashes map[LibraryFile]libSrcData
}

// IsZero reports whether json's omitzero option should omit this value.
func (s Source) IsZero() bool {
	return s.TID == "" || s.Library == "" || s.Path == ""
}

// MarshalJSONTo implements json.MarshalerTo. It normalizes the path to forward slashes so that sources are always
// written in a platform-independent form.
func (s Source) MarshalJSONTo(enc *jsontext.Encoder) error {
	type alias Source
	a := alias(s)
	a.Path = strings.ReplaceAll(a.Path, "\\", "/")
	return json.MarshalEncode(enc, &a)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. It normalizes the path to forward slashes so that sources written
// on Windows, with backslash separators, remain usable on other platforms.
func (s *Source) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	type alias Source
	var a alias
	if err := json.UnmarshalDecode(dec, &a); err != nil {
		return err
	}
	*s = Source(a)
	s.Path = strings.ReplaceAll(s.Path, "\\", "/")
	return nil
}

func (s Source) collectInto(m map[LibraryFile]struct{}) {
	if !s.IsZero() {
		if _, exists := m[s.LibraryFile]; !exists {
			m[s.LibraryFile] = struct{}{}
		}
	}
}

func (s Source) String() string {
	return s.LibraryFile.String() + "\n" + i18n.Text("ID: ") + string(s.TID)
}

func (l LibraryFile) String() string {
	return i18n.Text("Library: ") + l.Library + "\n" + i18n.Text("Path: ") + l.Path
}

// PrepareHashes loads, or reloads if modified, the hashes of the library files the provider's nodes are sourced from. A
// nil matcher does nothing.
func (sm *SrcMatcher) PrepareHashes(provider ListProvider) {
	if sm == nil {
		return
	}
	neededLibs := make(map[LibraryFile]struct{})
	forEachSourcedNode(provider, func(node sourcedNode) { node.GetSource().collectInto(neededLibs) })
	for libFile := range neededLibs {
		sm.prepareHashesFor(libFile)
	}
}

// prepareHashesFor loads the hashes of the library file, reloads them if the file was modified or its library moved,
// and drops them if the file or its library is gone.
func (sm *SrcMatcher) prepareHashesFor(libFile LibraryFile) {
	lib := GlobalSettings().Libraries.Lookup(libFile.Library)
	if lib == nil {
		delete(sm.libHashes, libFile)
		return
	}
	if sm.libHashes == nil {
		sm.libHashes = make(map[LibraryFile]libSrcData)
	}
	p := filepath.Join(lib.Path(), filepath.FromSlash(libFile.Path))
	stat, err := os.Stat(p)
	if err != nil {
		delete(sm.libHashes, libFile)
		return
	}
	srcData := libSrcData{
		path:       p,
		timestamp:  stat.ModTime(),
		size:       stat.Size(),
		dataHashes: make(map[tid.TID]HashAndData),
	}
	if data, exists := sm.libHashes[libFile]; exists {
		if data.path == p && data.timestamp.Equal(srcData.timestamp) && data.size == srcData.size {
			return // We've already loaded this exact version of the file.
		}
		delete(sm.libHashes, libFile)
	}
	dir := os.DirFS(filepath.Dir(p))
	file := filepath.Base(p)
	fi := FileInfoFor(p)
	if fi == nil || len(fi.UTI.Extensions) == 0 {
		return
	}
	switch fi.UTI.Extensions[0] {
	case TraitsExt:
		var data []*Trait
		if data, err = NewTraitsFromFile(dir, file); err == nil {
			NodesToHashesByID(srcData.dataHashes, data...)
			Traverse(func(t *Trait) bool {
				NodesToHashesByID(srcData.dataHashes, t.Modifiers...)
				return false
			}, false, false, data...)
		}
	case TraitModifiersExt:
		var data []*TraitModifier
		if data, err = NewTraitModifiersFromFile(dir, file); err == nil {
			NodesToHashesByID(srcData.dataHashes, data...)
		}
	case SkillsExt:
		var data []*Skill
		if data, err = NewSkillsFromFile(dir, file); err == nil {
			NodesToHashesByID(srcData.dataHashes, data...)
		}
	case SpellsExt:
		var data []*Spell
		if data, err = NewSpellsFromFile(dir, file); err == nil {
			NodesToHashesByID(srcData.dataHashes, data...)
		}
	case EquipmentExt:
		var data []*Equipment
		if data, err = NewEquipmentFromFile(dir, file); err == nil {
			NodesToHashesByID(srcData.dataHashes, data...)
			Traverse(func(e *Equipment) bool {
				NodesToHashesByID(srcData.dataHashes, e.Modifiers...)
				return false
			}, false, false, data...)
		}
	case EquipmentModifiersExt:
		var data []*EquipmentModifier
		if data, err = NewEquipmentModifiersFromFile(dir, file); err == nil {
			NodesToHashesByID(srcData.dataHashes, data...)
		}
	case NotesExt:
		var data []*Note
		if data, err = NewNotesFromFile(dir, file); err == nil {
			NodesToHashesByID(srcData.dataHashes, data...)
		}
	}
	sm.libHashes[libFile] = srcData
}

// Match returns the source state of the given data, along with the library's copy of it when one was found. A nil
// matcher finds nothing.
func (sm *SrcMatcher) Match(data SrcProvider) (state srcstate.Value, match any) {
	src := data.GetSource()
	if src.IsZero() {
		return srcstate.Custom, nil
	}
	if sm == nil {
		return srcstate.Missing, nil
	}
	if srcData, ok := sm.libHashes[src.LibraryFile]; ok {
		var dataHash HashAndData
		if dataHash, ok = srcData.dataHashes[src.TID]; ok {
			if dataHash.Hash == Hash64(data) {
				return srcstate.Matched, dataHash.Data
			}
			return srcstate.Mismatched, dataHash.Data
		}
	}
	return srcstate.Missing, nil
}

// maxUnownedSrcFiles is the most library files unownedSrcMatcher holds on to.
const maxUnownedSrcFiles = 8

var (
	// unownedSrcMatcher is the source matcher MatchSource uses for a node that has no data owner, or whose data owner
	// has no matcher. It lasts as long as the app, and each library file it holds comes with every node parsed from it,
	// so it holds on to only the files in unownedSrcFiles.
	unownedSrcMatcher SrcMatcher
	// unownedSrcFiles lists the library files unownedSrcMatcher was last used for, the latest last.
	unownedSrcFiles []LibraryFile
)

// useUnownedSrcFile records the library file as the latest unownedSrcMatcher is used for, and drops the files beyond
// maxUnownedSrcFiles.
func useUnownedSrcFile(libFile LibraryFile) {
	unownedSrcFiles = append(slices.DeleteFunc(unownedSrcFiles, func(one LibraryFile) bool { return one == libFile }),
		libFile)
	if excess := len(unownedSrcFiles) - maxUnownedSrcFiles; excess > 0 {
		unownedSrcFiles = slices.Delete(unownedSrcFiles, 0, excess)
	}
	maps.DeleteFunc(unownedSrcMatcher.libHashes, func(one LibraryFile, _ libSrcData) bool {
		return !slices.Contains(unownedSrcFiles, one)
	})
}

// MatchSource returns the source state of the node, along with the library's copy of it when one was found. It uses the
// source matcher of the node's data owner, or a shared one when there is none, first loading the node's library file,
// or reloading it if modified, since the matcher may not have been prepared with it.
func MatchSource[T Node[T]](node T) (state srcstate.Value, match any) {
	var sm *SrcMatcher
	if owner := node.DataOwner(); !xreflect.IsNil(owner) {
		sm = owner.SourceMatcher()
	}
	if src := node.GetSource(); !src.IsZero() {
		if sm == nil {
			useUnownedSrcFile(src.LibraryFile)
			sm = &unownedSrcMatcher
		}
		sm.prepareHashesFor(src.LibraryFile)
	}
	return sm.Match(node)
}

// GetSource returns the source of this data.
func (s *SourcedID) GetSource() Source {
	return s.Source
}

// ClearSource clears the source of this data.
func (s *SourcedID) ClearSource() {
	s.Source = Source{}
}

// SetSource sets the source of this data.
func (s *SourcedID) SetSource(src Source) {
	s.Source = src
}

// AdjustSource sets TID and Source from original according to mode: a Copy keeps original's TID, and a Reference to an
// original with no Source of its own points Source at original's TID in from; otherwise original's Source is copied.
func (s *SourcedID) AdjustSource(from LibraryFile, original SourcedID, mode CloneMode) {
	if mode == Copy {
		s.TID = original.TID
	}
	if mode == Reference && original.Source.Library == "" {
		s.Source = Source{LibraryFile: from, TID: original.TID}
	} else {
		s.Source = original.Source
	}
}
