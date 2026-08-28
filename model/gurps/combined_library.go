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
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
)

// CombinedLibrarySource identifies one folder whose data files contribute to a combined library. The order sources
// appear in CombinedLibraryOptions.Sources is their priority: when two sources provide the same item, the one from the
// earlier source keeps its stats, while the later one contributes only its page references.
type CombinedLibrarySource struct {
	Library *Library
	Folder  string // relative to the library root, using slashes
}

// CombinedLibraryOptions holds the options for CreateCombinedLibrary.
type CombinedLibraryOptions struct {
	Name              string
	Sources           []CombinedLibrarySource // highest priority first
	IncludeSubfolders bool
}

// combinedLibraryFileTypes lists the extensions that participate in combining, in the order their combined files are
// produced, along with the suffix used to name each combined file (e.g. "Combo Equipment.eqp").
var combinedLibraryFileTypes = []struct {
	ext    string
	suffix func() string
}{
	{ext: TraitsExt, suffix: func() string { return i18n.Text("Traits") }},
	{ext: TraitModifiersExt, suffix: func() string { return i18n.Text("Trait Modifiers") }},
	{ext: SkillsExt, suffix: func() string { return i18n.Text("Skills") }},
	{ext: SpellsExt, suffix: func() string { return i18n.Text("Spells") }},
	{ext: EquipmentExt, suffix: func() string { return i18n.Text("Equipment") }},
	{ext: EquipmentModifiersExt, suffix: func() string { return i18n.Text("Equipment Modifiers") }},
	{ext: NotesExt, suffix: func() string { return i18n.Text("Rules") }},
}

// combinedSourceFile is one data file that will contribute to a combined file, expressed both as the LibraryFile that
// the combined rows record as their source and as the absolute path used to read it. The source folders themselves are
// only ever read, never written: everything produced lands in the destination directory handed to
// CreateCombinedLibrary.
type combinedSourceFile struct {
	libFile LibraryFile
	from    *Library
}

func (f combinedSourceFile) absPath() string {
	return filepath.Join(f.from.Path(), filepath.FromSlash(f.libFile.Path))
}

// CreateCombinedLibrary combines the like-for-like data files of the source folders into one file per data type,
// written into destDir (typically a new folder within the user library), and returns the paths of the files it wrote.
// The source folders are treated as read-only. Within each combined file, items that appear in more than one source
// are collapsed into a single entry that keeps the stats of the highest-priority source and the union of every
// source's page references. Identity is the item's kind plus its name -- plus the tech level (and, for skills, the
// specialization) where the type has one, since items like "Assault Boots" legitimately exist once per TL.
func CreateCombinedLibrary(opts CombinedLibraryOptions, destDir string) ([]string, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return nil, errs.New(i18n.Text("A name for the combined library must be provided"))
	}
	if len(opts.Sources) == 0 {
		return nil, errs.New(i18n.Text("At least one source folder must be provided"))
	}
	// Collect every contributing file up front, before anything is written, so that a destination that happens to
	// overlap a source folder can never feed this run's own output back into itself.
	filesByExt := make(map[string][]combinedSourceFile)
	for _, src := range opts.Sources {
		if err := collectCombinedSourceFiles(src, opts.IncludeSubfolders, filesByExt); err != nil {
			return nil, err
		}
	}
	var created []string
	for _, ft := range combinedLibraryFileTypes {
		files := filesByExt[ft.ext]
		if len(files) == 0 {
			continue
		}
		target := filepath.Join(destDir, name+" "+ft.suffix()+ft.ext)
		if err := combineFilesInto(ft.ext, files, target); err != nil {
			return nil, err
		}
		created = append(created, target)
	}
	if len(created) == 0 {
		return nil, errs.New(i18n.Text("The selected folders contain no library data files to combine"))
	}
	return created, nil
}

// CombinedLibraryFilePaths returns the file paths that CreateCombinedLibrary would write into destDir for the given
// name, one per combinable data type. Only those whose data types are actually present among the sources end up being
// written.
func CombinedLibraryFilePaths(name, destDir string) []string {
	paths := make([]string, 0, len(combinedLibraryFileTypes))
	for _, ft := range combinedLibraryFileTypes {
		paths = append(paths, filepath.Join(destDir, name+" "+ft.suffix()+ft.ext))
	}
	return paths
}

// collectCombinedSourceFiles gathers the combinable data files of one source folder into filesByExt, in a stable
// order.
func collectCombinedSourceFiles(src CombinedLibrarySource, includeSubfolders bool, filesByExt map[string][]combinedSourceFile) error {
	combinable := make(map[string]bool, len(combinedLibraryFileTypes))
	for _, ft := range combinedLibraryFileTypes {
		combinable[ft.ext] = true
	}
	folder := path.Clean(filepath.ToSlash(src.Folder))
	dirOnDisk := filepath.Join(src.Library.Path(), filepath.FromSlash(folder))
	add := func(relToFolder string) {
		ext := strings.ToLower(path.Ext(relToFolder))
		if combinable[ext] {
			filesByExt[ext] = append(filesByExt[ext], combinedSourceFile{
				libFile: LibraryFile{
					Library: src.Library.Key(),
					Path:    path.Join(folder, relToFolder),
				},
				from: src.Library,
			})
		}
	}
	if includeSubfolders {
		// fs.WalkDir visits entries in lexical order, so the result is already deterministic.
		return errs.Wrap(fs.WalkDir(os.DirFS(dirOnDisk), ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.HasPrefix(d.Name(), ".") {
				if d.IsDir() && p != "." {
					return fs.SkipDir
				}
				return nil
			}
			if !d.IsDir() {
				add(p)
			}
			return nil
		}))
	}
	entries, err := os.ReadDir(dirOnDisk)
	if err != nil {
		return errs.Wrap(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			names = append(names, entry.Name())
		}
	}
	xstrings.SortStringsNaturalAscending(names)
	for _, one := range names {
		add(one)
	}
	return nil
}

// combineFilesInto loads, merges and writes one data type's combined file.
func combineFilesInto(ext string, files []combinedSourceFile, target string) error {
	switch ext {
	case TraitsExt:
		return combineTypedFilesInto(files, target, NewTraitsFromFile, SaveTraits, func(t *Trait) string {
			return combinedKey(t.Kind(), t.Name)
		})
	case TraitModifiersExt:
		return combineTypedFilesInto(files, target, NewTraitModifiersFromFile, SaveTraitModifiers,
			func(m *TraitModifier) string { return combinedKey(m.Kind(), m.Name) })
	case SkillsExt:
		return combineTypedFilesInto(files, target, NewSkillsFromFile, SaveSkills, func(s *Skill) string {
			var tl string
			if s.TechLevel != nil {
				tl = *s.TechLevel
			}
			return combinedKey(s.Kind(), s.Name, s.Specialization, tl)
		})
	case SpellsExt:
		return combineTypedFilesInto(files, target, NewSpellsFromFile, SaveSpells, func(s *Spell) string {
			var tl string
			if s.TechLevel != nil {
				tl = *s.TechLevel
			}
			return combinedKey(s.Kind(), s.Name, tl)
		})
	case EquipmentExt:
		return combineTypedFilesInto(files, target, NewEquipmentFromFile, SaveEquipment, func(e *Equipment) string {
			return combinedKey(e.Kind(), e.Name, e.TechLevel)
		})
	case EquipmentModifiersExt:
		return combineTypedFilesInto(files, target, NewEquipmentModifiersFromFile, SaveEquipmentModifiers,
			func(m *EquipmentModifier) string { return combinedKey(m.Kind(), m.Name, m.TechLevel) })
	case NotesExt:
		return combineTypedFilesInto(files, target, NewNotesFromFile, SaveNotes, func(n *Note) string {
			return combinedKey(n.Kind(), n.MarkDown)
		})
	default:
		return errs.Newf("unsupported extension for combining: %s", ext)
	}
}

// combineTypedFilesInto merges the rows of the given files, in priority order, and writes the result to target. Each
// row is cloned in Reference mode, so the combined rows point back at the file they came from and remain usable with
// the source-sync machinery, while the files being read are never modified.
func combineTypedFilesInto[T Node[T]](files []combinedSourceFile, target string,
	load func(fs.FS, string) ([]T, error), save func([]T, string) error, key func(T) string,
) error {
	var combined []T
	for _, f := range files {
		p := f.absPath()
		rows, err := load(os.DirFS(filepath.Dir(p)), filepath.Base(p))
		if err != nil {
			return errs.NewWithCause(p, err)
		}
		var noParent T
		clones := make([]T, 0, len(rows))
		for _, row := range rows {
			clones = append(clones, row.Clone(f.libFile, nil, noParent, Reference))
		}
		combined = mergeCombinedRows(combined, clones, key)
	}
	return save(combined, target)
}

// mergeCombinedRows folds incoming into existing. A row whose key matches an existing row's contributes its page
// references to it -- and, when both are containers, has its children merged recursively -- while everything else is
// appended. incoming must be lower priority than everything already in existing, since a key match always keeps the
// existing row's stats.
func mergeCombinedRows[T Node[T]](existing, incoming []T, key func(T) string) []T {
	lookup := make(map[string]T, len(existing))
	for _, one := range existing {
		k := key(one)
		if _, ok := lookup[k]; !ok {
			lookup[k] = one
		}
	}
	for _, in := range incoming {
		k := key(in)
		match, ok := lookup[k]
		if !ok {
			var noParent T
			in.SetParent(noParent)
			existing = append(existing, in)
			lookup[k] = in
			continue
		}
		mergeCombinedPageRefs(match, in)
		if isCombinedContainer(match) && isCombinedContainer(in) {
			children := mergeCombinedRows(match.NodeChildren(), in.NodeChildren(), key)
			for _, child := range children {
				child.SetParent(match)
			}
			match.SetChildren(children)
		}
	}
	return existing
}

// isCombinedContainer reports whether the node is a container. Every type that participates in combining has a
// Container() method on its concrete type; it just isn't part of the Node constraint.
func isCombinedContainer[T Node[T]](node T) bool {
	c, ok := any(node).(interface{ Container() bool })
	return ok && c.Container()
}

// mergeCombinedPageRefs adds src's page references to dst's.
func mergeCombinedPageRefs[T Node[T]](dst, src T) {
	switch d := any(dst).(type) {
	case *Trait:
		if s, ok := any(src).(*Trait); ok {
			d.PageRef = combinePageRefs(d.PageRef, s.PageRef)
		}
	case *TraitModifier:
		if s, ok := any(src).(*TraitModifier); ok {
			d.PageRef = combinePageRefs(d.PageRef, s.PageRef)
		}
	case *Skill:
		if s, ok := any(src).(*Skill); ok {
			d.PageRef = combinePageRefs(d.PageRef, s.PageRef)
		}
	case *Spell:
		if s, ok := any(src).(*Spell); ok {
			d.PageRef = combinePageRefs(d.PageRef, s.PageRef)
		}
	case *Equipment:
		if s, ok := any(src).(*Equipment); ok {
			d.PageRef = combinePageRefs(d.PageRef, s.PageRef)
		}
	case *EquipmentModifier:
		if s, ok := any(src).(*EquipmentModifier); ok {
			d.PageRef = combinePageRefs(d.PageRef, s.PageRef)
		}
	case *Note:
		if s, ok := any(src).(*Note); ok {
			d.PageRef = combinePageRefs(d.PageRef, s.PageRef)
		}
	}
}

// combinePageRefs merges two comma-separated page reference lists, keeping a's entries first and dropping duplicates.
func combinePageRefs(a, b string) string {
	if strings.TrimSpace(b) == "" {
		return a
	}
	if strings.TrimSpace(a) == "" {
		return b
	}
	refs := strings.Split(a, ",")
	for i, ref := range refs {
		refs[i] = strings.TrimSpace(ref)
	}
	for one := range strings.SplitSeq(b, ",") {
		one = strings.TrimSpace(one)
		if one == "" {
			continue
		}
		exists := false
		for _, ref := range refs {
			if strings.EqualFold(ref, one) {
				exists = true
				break
			}
		}
		if !exists {
			refs = append(refs, one)
		}
	}
	return strings.Join(refs, ",")
}

// combinedKey builds the identity key used to detect collisions while combining. Keys are case-insensitive, since the
// same item may not be capitalized identically in every book.
func combinedKey(parts ...string) string {
	for i, part := range parts {
		parts[i] = strings.ToLower(strings.TrimSpace(part))
	}
	return strings.Join(parts, "\x00")
}
