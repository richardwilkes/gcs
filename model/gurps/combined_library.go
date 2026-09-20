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
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xos"
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

// combinedLibraryExts lists the extensions that participate in combining, in the order their combined files are
// produced.
var combinedLibraryExts = []string{
	TraitsExt,
	TraitModifiersExt,
	SkillsExt,
	SpellsExt,
	EquipmentExt,
	EquipmentModifiersExt,
	NotesExt,
}

// combinedStandardTables lists, per extension, the standard table names used across libraries. A file whose name ends
// with one of these belongs to that table no matter what precedes it, which is how "Low Tech Companion 2 Equipment
// Modifiers" and "Basic Set Equipment Modifiers" end up in the same combined file by default. Longer names must come
// before any name that is a suffix of theirs. These match the data files' own naming convention, so they are
// deliberately not translated: translating them would stop the names matching the files.
var combinedStandardTables = map[string][]string{
	TraitsExt:             {"Traits"},
	TraitModifiersExt:     {"Enhancement Modifiers", "Limitation Modifiers", "Trait Modifiers"},
	SkillsExt:             {"Skills"},
	SpellsExt:             {"Spells"},
	EquipmentExt:          {"Equipment"},
	EquipmentModifiersExt: {"Equipment Modifiers", "TL Cost Modifiers"},
	NotesExt:              {"Rules", "Notes"},
}

// CombinableExt reports whether files with the given extension participate in combining.
func CombinableExt(ext string) bool {
	return slices.Contains(combinedLibraryExts, strings.ToLower(ext))
}

// combinedTempSuffix is appended to a combined file's name while it is being staged. Staged files are only renamed
// into place once every one of them has been written successfully.
const combinedTempSuffix = ".combining"

// combinedSourceFile is one data file that will contribute to a combined file, expressed both as the LibraryFile that
// identifies it within its library and as the absolute path used to read it. The source folders themselves are only
// ever read, never written: everything produced lands in the destination directory handed to CombinedPlan.Create.
type combinedSourceFile struct {
	libFile LibraryFile
	from    *Library
}

func (f combinedSourceFile) absPath() string {
	return filepath.Join(f.from.Path(), filepath.FromSlash(f.libFile.Path))
}

// combinedSourceGroup is a run of staged files whose rows are all kept when they are merged into a combined file,
// even when they share an identity: a book that lists the same item twice on purpose still lists it twice after
// combining. Matches between different groups collapse instead, so a group is one book's contribution.
type combinedSourceGroup struct {
	files []combinedSourceFile
}

// CombinedComponent is one source data file staged to contribute to a combined file. Path is relative to the
// library's root, in slash form, and its first element is the source folder the component came from.
type CombinedComponent struct {
	Library *Library
	Path    string
}

// Ext returns the component's file extension, lower-cased.
func (c CombinedComponent) Ext() string {
	return strings.ToLower(path.Ext(c.Path))
}

// same reports whether two components identify the same file. Paths are compared without regard to case, since the
// file systems GCS libraries usually live on do not distinguish it.
func (c CombinedComponent) same(other CombinedComponent) bool {
	return c.Library.Key() == other.Library.Key() && strings.EqualFold(c.Path, other.Path)
}

// book identifies the source folder the component came from, so that consecutive components from the same book can be
// treated as that book's single contribution.
func (c CombinedComponent) book() string {
	folder, _, _ := strings.Cut(c.Path, "/")
	return c.Library.Key() + "\x00" + strings.ToLower(folder)
}

// CombinedFile is one output file of a combination: the components that merge into it, highest priority first, plus
// where it lands and what it is called. Ext is fixed at staging time; only components with the same extension may be
// staged into the file. Until CustomName is set, the file is named "<plan name> <Table><Ext>", tracking the plan's
// name; renaming the file sets CustomName and pins it.
type CombinedFile struct {
	Subpath    string // output subfolder, slash form, "" for the top level
	Table      string // the table the file was staged for; used both for default naming and to route later additions
	CustomName string // overrides the default name when set; never includes the extension
	Ext        string
	Components []CombinedComponent
}

// FileName returns the output file's name for a combination with the given plan name.
func (f *CombinedFile) FileName(planName string) string {
	if f.CustomName != "" {
		return f.CustomName + f.Ext
	}
	if f.Table != "" {
		return planName + " " + f.Table + f.Ext
	}
	return planName + f.Ext
}

// CombinedPlan is the staging area for a combination: every output file it will produce and the components each is
// built from. AddSource and AddFile stage components at their default locations; the slices may then be freely
// rearranged -- components moved between files of the same extension, reordered, or removed, files renamed or created
// -- before Create writes the result.
type CombinedPlan struct {
	Files []*CombinedFile
}

// AddSource stages every combinable data file of the given source folder at its default location, skipping files that
// are already staged. New components are appended below existing ones, so sources added earlier keep priority.
func (p *CombinedPlan) AddSource(src CombinedLibrarySource, includeSubfolders bool) error {
	perSource, err := collectCombinedSourceFiles(src, includeSubfolders)
	if err != nil {
		return err
	}
	for _, id := range perSource.ids {
		f := p.ensureFile(id.key.subpath, id.key.ext, id.display)
		for _, one := range perSource.files[id.key] {
			comp := CombinedComponent{Library: one.from, Path: one.libFile.Path}
			if !p.Contains(comp) {
				f.Components = append(f.Components, comp)
			}
		}
	}
	return nil
}

// AddFile stages a single data file at its default location, reporting whether it was added. A file that is not a
// combinable type, or is already staged, is left alone. relPath is relative to the source folder, in slash form.
func (p *CombinedPlan) AddFile(src CombinedLibrarySource, relPath string) bool {
	folder := path.Clean(filepath.ToSlash(src.Folder))
	relPath = path.Clean(filepath.ToSlash(relPath))
	ext := strings.ToLower(path.Ext(relPath))
	if !CombinableExt(ext) {
		return false
	}
	comp := CombinedComponent{Library: src.Library, Path: path.Join(folder, relPath)}
	if p.Contains(comp) {
		return false
	}
	subpath := path.Dir(relPath)
	if subpath == "." {
		subpath = ""
	}
	display := combinedTableName(strings.TrimSuffix(path.Base(relPath), path.Ext(relPath)), ext,
		append([]string{path.Base(folder)}, strings.Split(subpath, "/")...))
	f := p.ensureFile(subpath, ext, display)
	f.Components = append(f.Components, comp)
	return true
}

// Contains reports whether the component is staged anywhere in the plan.
func (p *CombinedPlan) Contains(comp CombinedComponent) bool {
	for _, f := range p.Files {
		for _, one := range f.Components {
			if one.same(comp) {
				return true
			}
		}
	}
	return false
}

// HasContent reports whether any file has at least one component, which is what Create requires.
func (p *CombinedPlan) HasContent() bool {
	for _, f := range p.Files {
		if len(f.Components) != 0 {
			return true
		}
	}
	return false
}

// DropEmptyDefaultFiles removes files that have no components left and were never renamed. Files the user named are
// kept, since an empty one is presumably about to be filled.
func (p *CombinedPlan) DropEmptyDefaultFiles() {
	p.Files = slices.DeleteFunc(p.Files, func(f *CombinedFile) bool {
		return len(f.Components) == 0 && f.CustomName == ""
	})
}

// ensureFile returns the staged file for the given table, creating it if the plan does not have it yet. Only files
// staged for a table are candidates: a file the user created from scratch has no table and never captures later
// additions.
func (p *CombinedPlan) ensureFile(subpath, ext, table string) *CombinedFile {
	for _, f := range p.Files {
		if f.Subpath == subpath && f.Ext == ext && f.Table != "" && strings.EqualFold(f.Table, table) {
			return f
		}
	}
	f := &CombinedFile{Subpath: subpath, Table: table, Ext: ext}
	p.Files = append(p.Files, f)
	return f
}

// sortedFiles returns the plan's files that have components, in the deterministic order Create writes them and
// Targets lists them: by subfolder, then data type, then file name.
func (p *CombinedPlan) sortedFiles(name string) []*CombinedFile {
	files := make([]*CombinedFile, 0, len(p.Files))
	for _, f := range p.Files {
		if len(f.Components) != 0 {
			files = append(files, f)
		}
	}
	extOrder := make(map[string]int, len(combinedLibraryExts))
	for i, ext := range combinedLibraryExts {
		extOrder[ext] = i
	}
	slices.SortStableFunc(files, func(a, b *CombinedFile) int {
		if c := xstrings.NaturalCmp(a.Subpath, b.Subpath, true); c != 0 {
			return c
		}
		if c := extOrder[a.Ext] - extOrder[b.Ext]; c != 0 {
			return c
		}
		return xstrings.NaturalCmp(a.FileName(name), b.FileName(name), true)
	})
	return files
}

// Targets returns the file paths Create would write into destDir for the given plan name. Files without components
// produce nothing and are not listed.
func (p *CombinedPlan) Targets(name, destDir string) []string {
	files := p.sortedFiles(name)
	targets := make([]string, 0, len(files))
	for _, f := range files {
		targets = append(targets, filepath.Join(destDir, filepath.FromSlash(f.Subpath), f.FileName(name)))
	}
	return targets
}

// Create writes the plan's combined files into destDir (typically a new folder within the user library) and returns
// the paths of the files it wrote. The staged source files are treated as read-only. Files without components are
// skipped.
//
// Within each combined file, items that appear in more than one component are collapsed into a single entry that
// keeps the stats of the highest-priority component and the union of every component's page references, while
// consecutive components from the same book are that book's single contribution, whose duplicates all survive.
// Identity is the item's name -- plus the tech level (and, for skills, the specialization) where the type has one,
// since items like "Assault Boots" legitimately exist once per TL.
//
// The rows written are root data: they carry no Source of their own, which makes the combined file the source of
// truth for what it contains. A row dragged from a combined file onto a sheet therefore anchors to the combined file,
// so "Sync with Source" restores the merged stats and merged page references rather than the originating book's. To
// pick up later changes to a book, re-run the combination; rows that still match keep the IDs they had, so sheets
// referring to the combined file survive the regeneration.
func (p *CombinedPlan) Create(name, destDir string) ([]string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errs.New(i18n.Text("A name for the combined library must be provided"))
	}
	files := p.sortedFiles(name)
	if len(files) == 0 {
		return nil, errs.New(i18n.Text("There are no data files staged to combine"))
	}
	// Load and merge everything before writing a single byte, so that a failure part way through cannot leave a
	// half-written folder behind.
	seen := make(map[string]bool, len(files))
	pending := make([]pendingCombinedFile, 0, len(files))
	for _, f := range files {
		for _, comp := range f.Components {
			if comp.Ext() != f.Ext {
				return nil, errs.Newf("component %s cannot belong to a %s file", comp.Path, f.Ext)
			}
		}
		target := filepath.Join(destDir, filepath.FromSlash(f.Subpath), f.FileName(name))
		lower := strings.ToLower(target)
		if seen[lower] {
			return nil, errs.Newf(i18n.Text("Two combined files would be written to the same path: %s"), target)
		}
		seen[lower] = true
		save, err := prepareCombinedFile(f.Ext, batchCombinedComponents(f.Components), target)
		if err != nil {
			return nil, err
		}
		pending = append(pending, pendingCombinedFile{target: target, save: save})
	}
	return writeCombinedFiles(destDir, pending)
}

// batchCombinedComponents folds runs of consecutive components that share a book into single groups, so that a book's
// files contribute as one and their deliberate duplicates survive, while components the user has interleaved from
// different books merge in the priority order shown.
func batchCombinedComponents(components []CombinedComponent) []combinedSourceGroup {
	var groups []combinedSourceGroup
	var lastBook string
	for _, comp := range components {
		src := combinedSourceFile{
			libFile: LibraryFile{Library: comp.Library.Key(), Path: comp.Path},
			from:    comp.Library,
		}
		if book := comp.book(); len(groups) == 0 || book != lastBook {
			groups = append(groups, combinedSourceGroup{files: []combinedSourceFile{src}})
			lastBook = book
		} else {
			groups[len(groups)-1].files = append(groups[len(groups)-1].files, src)
		}
	}
	return groups
}

// CreateCombinedLibrary stages the given source folders at their default locations and writes the result into
// destDir: the outcome of accepting the staging as-is, with no manual adjustments. See CombinedPlan.Create for the
// semantics.
func CreateCombinedLibrary(opts CombinedLibraryOptions, destDir string) ([]string, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return nil, errs.New(i18n.Text("A name for the combined library must be provided"))
	}
	if len(opts.Sources) == 0 {
		return nil, errs.New(i18n.Text("At least one source folder must be provided"))
	}
	var plan CombinedPlan
	for _, src := range opts.Sources {
		if err := plan.AddSource(src, opts.IncludeSubfolders); err != nil {
			return nil, err
		}
	}
	if !plan.HasContent() {
		return nil, errs.New(i18n.Text("The selected folders contain no library data files to combine"))
	}
	return plan.Create(name, destDir)
}

// pendingCombinedFile is a combined file that has been fully loaded and merged in memory but not yet written.
type pendingCombinedFile struct {
	save   func(path string) error
	target string
}

// combinedTableKey identifies one default staging destination: a table of one data type within one subfolder. The
// table name is compared lower-cased, so that books that capitalize a table differently still share it.
type combinedTableKey struct {
	subpath string
	ext     string
	table   string
}

// combinedTableID pairs a table key with the display form of its table name, which the key deliberately leaves out.
type combinedTableID struct {
	key     combinedTableKey
	display string
}

// combinedSourceTables holds one source folder's combinable files, grouped into the tables they belong to. ids
// preserves the order the tables were first encountered in, so that iteration is deterministic.
type combinedSourceTables struct {
	ids   []combinedTableID
	files map[combinedTableKey][]combinedSourceFile
}

// collectCombinedSourceFiles gathers the combinable data files of one source folder, grouped into the tables they
// belong to. Within each table, files stay in their gathered order.
func collectCombinedSourceFiles(src CombinedLibrarySource, includeSubfolders bool) (*combinedSourceTables, error) {
	folder := path.Clean(filepath.ToSlash(src.Folder))
	dirOnDisk := filepath.Join(src.Library.Path(), filepath.FromSlash(folder))
	rel, err := gatherCombinableFiles(dirOnDisk, includeSubfolders)
	if err != nil {
		return nil, err
	}
	// One ordering rule for both cases. Turning "include subfolders" on adds files, but never reorders the ones that
	// were already there, so it cannot change the outcome for the files the two modes have in common.
	xstrings.SortStringsNaturalAscending(rel)
	perSource := &combinedSourceTables{files: make(map[combinedTableKey][]combinedSourceFile)}
	for _, one := range rel {
		ext := strings.ToLower(path.Ext(one))
		if !CombinableExt(ext) {
			continue
		}
		subpath := path.Dir(one)
		if subpath == "." {
			subpath = ""
		}
		display := combinedTableName(strings.TrimSuffix(path.Base(one), path.Ext(one)), ext,
			append([]string{path.Base(folder)}, strings.Split(subpath, "/")...))
		key := combinedTableKey{subpath: subpath, ext: ext, table: strings.ToLower(display)}
		if _, ok := perSource.files[key]; !ok {
			perSource.ids = append(perSource.ids, combinedTableID{key: key, display: display})
		}
		perSource.files[key] = append(perSource.files[key], combinedSourceFile{
			libFile: LibraryFile{
				Library: src.Library.Key(),
				Path:    path.Join(folder, one),
			},
			from: src.Library,
		})
	}
	return perSource, nil
}

// combinedTableName determines which table a data file belongs to by default from its name (without the extension). A
// name that ends with one of the extension's standard table names belongs to that table, no matter what book prefix
// precedes it. Otherwise the name is a specialized list: the folder or subfolder name it starts with is stripped, so
// that "Low Tech Armor (by location)" becomes "Armor (by location)", and what remains names the table. Both checks
// insist on a space boundary, so "Meta-Traits" is a table of its own rather than part of "Traits".
func combinedTableName(base, ext string, prefixes []string) string {
	for _, std := range combinedStandardTables[ext] {
		if strings.EqualFold(base, std) {
			return std
		}
		if len(base) > len(std)+1 && base[len(base)-len(std)-1] == ' ' &&
			strings.EqualFold(base[len(base)-len(std):], std) {
			return std
		}
	}
	// Longest first, so that a subfolder such as "Dungeon Fantasy 1" wins over its parent "Dungeon Fantasy".
	prefixes = slices.Clone(prefixes)
	slices.SortFunc(prefixes, func(a, b string) int { return len(b) - len(a) })
	for _, prefix := range prefixes {
		if len(base) > len(prefix)+1 && base[len(prefix)] == ' ' && strings.EqualFold(base[:len(prefix)], prefix) {
			// A separator may follow the prefix, as in "Dungeon Fantasy Adventure 1 - Mirror of the Fire
			// Demons Contents"; it belongs to the prefix, not the table name.
			if remainder := strings.TrimLeft(base[len(prefix)+1:], "-– "); remainder != "" {
				return remainder
			}
			return base[len(prefix)+1:]
		}
	}
	return base
}

// writeCombinedFiles stages every prepared file alongside its target and only then renames them into place. The
// realistic failures -- unreadable source data, a full disk -- all strike before the first rename and leave the
// destination exactly as it was. A rename failing partway through the final loop would still leave a mix of old and
// new files, but renames within a directory that was just written to have no ordinary way to fail.
func writeCombinedFiles(destDir string, pending []pendingCombinedFile) ([]string, error) {
	dirExisted := xos.IsDir(destDir)
	dirs := combinedTargetDirs(pending)
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, errs.Wrap(err)
		}
	}
	temps := make([]string, len(pending))
	discardStaged := func() {
		for _, tmp := range temps {
			if tmp != "" {
				_ = os.Remove(tmp) //nolint:errcheck // Best-effort cleanup of a staged file
			}
		}
		// Deepest first, so that emptied subfolders unchain. Remove only succeeds while a directory is empty,
		// which is exactly when it should be removed.
		for _, dir := range slices.Backward(dirs) {
			_ = os.Remove(dir) //nolint:errcheck // Best-effort cleanup of directories this run created
		}
		if !dirExisted {
			_ = os.Remove(destDir) //nolint:errcheck // Best-effort cleanup of a directory this run created
		}
	}
	for i, p := range pending {
		tmp := p.target + combinedTempSuffix
		if err := p.save(tmp); err != nil {
			discardStaged()
			return nil, err
		}
		temps[i] = tmp
	}
	created := make([]string, 0, len(pending))
	for i, p := range pending {
		if err := os.Rename(temps[i], p.target); err != nil {
			discardStaged()
			return nil, errs.Wrap(err)
		}
		temps[i] = ""
		created = append(created, p.target)
	}
	return created, nil
}

// combinedTargetDirs returns the directories the pending files land in, sorted shallowest first and without
// duplicates.
func combinedTargetDirs(pending []pendingCombinedFile) []string {
	seen := make(map[string]bool)
	dirs := make([]string, 0, len(pending))
	for _, p := range pending {
		dir := filepath.Dir(p.target)
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	slices.SortFunc(dirs, func(a, b string) int { return len(a) - len(b) })
	return dirs
}

// gatherCombinableFiles returns the paths of the files within dirOnDisk, relative to it and using slashes. Hidden
// files and folders are skipped.
func gatherCombinableFiles(dirOnDisk string, includeSubfolders bool) ([]string, error) {
	var rel []string
	if includeSubfolders {
		if err := fs.WalkDir(os.DirFS(dirOnDisk), ".", func(p string, d fs.DirEntry, err error) error {
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
				rel = append(rel, p)
			}
			return nil
		}); err != nil {
			return nil, errs.Wrap(err)
		}
		return rel, nil
	}
	entries, err := os.ReadDir(dirOnDisk)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			rel = append(rel, entry.Name())
		}
	}
	return rel, nil
}

// prepareCombinedFile loads and merges one combined file's source groups, returning the function that writes the
// result.
func prepareCombinedFile(ext string, groups []combinedSourceGroup, target string) (func(string) error, error) {
	switch ext {
	case TraitsExt:
		return prepareTypedCombinedFile(groups, target, NewTraitsFromFile, SaveTraits, func(t *Trait) string {
			return combinedKey(combinedKeyKind(t.Container()), t.Name)
		})
	case TraitModifiersExt:
		return prepareTypedCombinedFile(groups, target, NewTraitModifiersFromFile, SaveTraitModifiers,
			func(m *TraitModifier) string { return combinedKey(combinedKeyKind(m.Container()), m.Name) })
	case SkillsExt:
		return prepareTypedCombinedFile(groups, target, NewSkillsFromFile, SaveSkills, func(s *Skill) string {
			var tl string
			if s.TechLevel != nil {
				tl = *s.TechLevel
			}
			return combinedKey(combinedKeyKind(s.Container()), s.Name, s.Specialization, tl)
		})
	case SpellsExt:
		return prepareTypedCombinedFile(groups, target, NewSpellsFromFile, SaveSpells, func(s *Spell) string {
			var tl string
			if s.TechLevel != nil {
				tl = *s.TechLevel
			}
			return combinedKey(combinedKeyKind(s.Container()), s.Name, tl)
		})
	case EquipmentExt:
		return prepareTypedCombinedFile(groups, target, NewEquipmentFromFile, SaveEquipment, func(e *Equipment) string {
			return combinedKey(combinedKeyKind(e.Container()), e.Name, e.TechLevel)
		})
	case EquipmentModifiersExt:
		return prepareTypedCombinedFile(groups, target, NewEquipmentModifiersFromFile, SaveEquipmentModifiers,
			func(m *EquipmentModifier) string {
				return combinedKey(combinedKeyKind(m.Container()), m.Name, m.TechLevel)
			})
	case NotesExt:
		return prepareTypedCombinedFile(groups, target, NewNotesFromFile, SaveNotes, func(n *Note) string {
			return combinedKey(combinedKeyKind(n.Container()), n.MarkDown)
		})
	default:
		return nil, errs.Newf("unsupported extension for combining: %s", ext)
	}
}

// prepareTypedCombinedFile merges the rows of the given source groups, in priority order, and returns the function
// that writes the result. The files being read are never modified.
func prepareTypedCombinedFile[T Node[T]](groups []combinedSourceGroup, target string,
	load func(fs.FS, string) ([]T, error), save func([]T, string) error, key func(T) string,
) (func(string) error, error) {
	var combined []T
	for _, group := range groups {
		var fromSource []T
		for _, f := range group.files {
			p := f.absPath()
			rows, err := load(os.DirFS(filepath.Dir(p)), filepath.Base(p))
			if err != nil {
				return nil, errs.NewWithCause(p, err)
			}
			var noParent T
			for _, row := range rows {
				// Duplicate rather than Reference: the combined file is the source of truth for what it contains,
				// so its rows are root data and a row dragged out of it anchors to the combined file. Duplicate
				// copies the original's Source verbatim, which is empty for book data but not for a source folder
				// that was itself derived, so clear it explicitly.
				clone := row.Clone(f.libFile, nil, noParent, Duplicate)
				clearCombinedSources(clone)
				fromSource = append(fromSource, clone)
			}
		}
		combined = mergeCombinedRows(combined, fromSource, key)
	}
	reuseCombinedIDs(combined, target, load, key)
	return func(p string) error { return save(combined, p) }, nil
}

// clearCombinedSources strips the Source from a row and everything beneath it. Node.ClearSource only affects the row
// it is called on.
func clearCombinedSources[T Node[T]](row T) {
	row.ClearSource()
	for _, child := range row.NodeChildren() {
		clearCombinedSources(child)
	}
}

// mergeCombinedRows folds incoming into existing. A row whose key matches an existing row's contributes its page
// references to it -- and, when both are containers, has its children merged recursively -- while everything else is
// appended. incoming must be lower priority than everything already in existing, since a key match always keeps the
// existing row's stats.
//
// incoming is one group's worth of rows, and rows appended from it are deliberately not added to the lookup: two
// rows from the same group that share a key are both kept, because a book is entitled to list an item twice. A
// later group's matching row merges into the first of them.
func mergeCombinedRows[T Node[T]](existing, incoming []T, key func(T) string) []T {
	lookup := make(map[string]T, len(existing))
	for _, one := range existing {
		k := key(one)
		if _, ok := lookup[k]; !ok {
			lookup[k] = one
		}
	}
	for _, in := range incoming {
		match, ok := lookup[key(in)]
		if !ok {
			var noParent T
			in.SetParent(noParent)
			existing = append(existing, in)
			continue
		}
		mergeCombinedPageRefs(match, in)
		if match.Container() && in.Container() {
			children := mergeCombinedRows(match.NodeChildren(), in.NodeChildren(), key)
			for _, child := range children {
				child.SetParent(match)
			}
			match.SetChildren(children)
		}
	}
	return existing
}

// tidSetter is satisfied by every node type, each of which embeds SourcedID. The TID field itself is not reachable
// through the Node constraint, so reusing IDs needs this.
type tidSetter interface {
	setTID(id tid.TID)
}

// combinedIDReuse carries the state of a single ID-reuse pass: the IDs recorded from the previous file, how many of
// each scoped identity have been handed out so far, and every ID already given away.
type combinedIDReuse struct {
	ids  map[string][]tid.TID
	next map[string]int
	used map[tid.TID]bool
}

// reuseCombinedIDs gives rows the IDs their counterparts had in a previously generated combined file, so that sheets
// and templates referring to the combined file keep working across regenerations. Rows with no counterpart keep the
// fresh IDs minted while cloning. This mirrors what the rules lookup download does with its own output.
func reuseCombinedIDs[T Node[T]](rows []T, target string, load func(fs.FS, string) ([]T, error), key func(T) string) {
	prior, err := load(os.DirFS(filepath.Dir(target)), filepath.Base(target))
	if err != nil {
		return // No previous file to take IDs from; the fresh ones stand.
	}
	reuse := &combinedIDReuse{
		ids:  make(map[string][]tid.TID),
		next: make(map[string]int),
		used: make(map[tid.TID]bool),
	}
	collectCombinedIDs(prior, "", reuse, key)
	if len(reuse.ids) == 0 {
		return
	}
	applyCombinedIDs(rows, "", reuse, key)
}

// combinedIDKey scopes a row's key by its ancestors', so that a child named X under one container cannot be confused
// with a child named X under another.
func combinedIDKey[T Node[T]](scope string, row T, key func(T) string) string {
	return scope + "\x01" + key(row)
}

// collectCombinedIDs records the IDs of a previously generated combined file, keyed by scoped identity. Siblings that
// share an identity are recorded in document order rather than collapsed, since a book is allowed to list the same
// item more than once and each of those rows needs an ID of its own to keep.
func collectCombinedIDs[T Node[T]](rows []T, scope string, reuse *combinedIDReuse, key func(T) string) {
	for _, row := range rows {
		k := combinedIDKey(scope, row, key)
		reuse.ids[k] = append(reuse.ids[k], row.ID())
		collectCombinedIDs(row.NodeChildren(), k, reuse, key)
	}
}

// applyCombinedIDs assigns previously recorded IDs to the rows that match them, pairing up same-identity siblings in
// document order. An ID is only ever handed out once, so rows cannot end up sharing one even if the previous file was
// malformed.
func applyCombinedIDs[T Node[T]](rows []T, scope string, reuse *combinedIDReuse, key func(T) string) {
	for _, row := range rows {
		k := combinedIDKey(scope, row, key)
		if available := reuse.ids[k]; reuse.next[k] < len(available) {
			id := available[reuse.next[k]]
			reuse.next[k]++
			if !reuse.used[id] {
				if setter, isSetter := any(row).(tidSetter); isSetter {
					setter.setTID(id)
					reuse.used[id] = true
				}
			}
		}
		applyCombinedIDs(row.NodeChildren(), k, reuse, key)
	}
}

// combinedKeyKind returns the first part of a row's identity key. Each data type is merged separately, so all the key
// has to distinguish is a container from a non-container of the same name. Node.Kind() would do that too, but its text
// is translated, and a key that shifts with the user's language would defeat ID reuse across regenerations.
func combinedKeyKind(container bool) string {
	if container {
		return "container"
	}
	return "item"
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
