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
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
)

// TestSourcePathSeparatorNormalization verifies that source paths are stored and round-tripped with forward slashes.
// See issue #1005: backslash paths written on Windows could not be located on other platforms, so the library source
// match status showed a question mark.
func TestSourcePathSeparatorNormalization(t *testing.T) {
	c := check.New(t)

	// A path loaded with backslash separators is normalized to forward slashes.
	const windowsAuthored = `{"id":"a","source":{"library":"Master Library","path":"Basic Set\\Basic Set Traits.adq","id":"x"}}`
	var loaded SourcedID
	c.NoError(jio.Unmarshal([]byte(windowsAuthored), &loaded), "backslash path should load")
	c.Equal("Basic Set/Basic Set Traits.adq", loaded.Source.Path, "path should be normalized on load")

	// An already-normalized path is left unchanged on load.
	const unixAuthored = `{"id":"a","source":{"library":"Master Library","path":"Basic Set/Basic Set Traits.adq","id":"x"}}`
	var loaded2 SourcedID
	c.NoError(jio.Unmarshal([]byte(unixAuthored), &loaded2), "forward-slash path should load")
	c.Equal("Basic Set/Basic Set Traits.adq", loaded2.Source.Path, "forward-slash path should be unchanged")

	// A source carrying a backslash path in memory is always written out with forward slashes.
	inMemory := SourcedID{
		TID: "a",
		Source: Source{
			LibraryFile: LibraryFile{Library: "Master Library", Path: `Basic Set\Basic Set Traits.adq`},
			TID:         "x",
		},
	}
	data, err := jio.Marshal(&inMemory)
	c.NoError(err, "source should marshal")
	c.False(strings.Contains(string(data), `\`), "marshaled output must not contain backslashes")
	c.True(strings.Contains(string(data), "Basic Set/Basic Set Traits.adq"), "marshaled output should use forward slashes")

	// Round-tripping the marshaled output preserves the normalized path.
	var roundTripped SourcedID
	c.NoError(jio.Unmarshal(data, &roundTripped), "marshaled source should load")
	c.Equal("Basic Set/Basic Set Traits.adq", roundTripped.Source.Path, "round-tripped path should remain normalized")
}

// TestForEachSourcedNodeVisitsEveryNodeOnce verifies that the walk shared by source syncing and hashing reaches every
// node a provider holds exactly once: nested children, the modifiers of traits and equipment, disabled nodes, and both
// equipment lists, while the lists a provider doesn't have (a loot sheet's traits, skills and spells) contribute
// nothing.
func TestForEachSourcedNodeVisitsEveryNodeOnce(t *testing.T) {
	c := check.New(t)
	visits := func(provider ListProvider) map[sourcedNode]int {
		m := make(map[sourcedNode]int)
		forEachSourcedNode(provider, func(node sourcedNode) { m[node]++ })
		return m
	}

	tmpl := NewTemplate()
	traitContainer := NewTrait(tmpl, nil, true)
	trait := NewTrait(tmpl, traitContainer, false)
	trait.Disabled = true
	traitMod := NewTraitModifier(tmpl, nil, false)
	trait.Modifiers = append(trait.Modifiers, traitMod)
	traitContainer.Children = append(traitContainer.Children, trait)
	tmpl.Traits = append(tmpl.Traits, traitContainer)
	skill := NewSkill(tmpl, nil, false)
	tmpl.Skills = append(tmpl.Skills, skill)
	spell := NewSpell(tmpl, nil, false)
	tmpl.Spells = append(tmpl.Spells, spell)
	eqp := NewEquipment(tmpl, nil, false)
	eqpMod := NewEquipmentModifier(tmpl, nil, false)
	eqp.Modifiers = append(eqp.Modifiers, eqpMod)
	tmpl.Equipment = append(tmpl.Equipment, eqp)
	note := NewNote(tmpl, nil, false)
	tmpl.Notes = append(tmpl.Notes, note)
	got := visits(tmpl)
	c.Equal(8, len(got), "every node of the template is visited")
	for _, node := range []sourcedNode{traitContainer, trait, traitMod, skill, spell, eqp, eqpMod, note} {
		c.Equal(1, got[node], "each node is visited exactly once")
	}

	e := NewEntity()
	before := len(visits(e))
	carried := NewEquipment(e, nil, false)
	e.CarriedEquipment = append(e.CarriedEquipment, carried)
	other := NewEquipment(e, nil, false)
	e.OtherEquipment = append(e.OtherEquipment, other)
	got = visits(e)
	c.Equal(before+2, len(got), "both equipment lists of an entity are visited")
	c.Equal(1, got[carried], "carried equipment is visited once")
	c.Equal(1, got[other], "other equipment is visited once")

	loot := NewLoot()
	lootEqp := NewEquipment(loot, nil, false)
	lootMod := NewEquipmentModifier(loot, nil, false)
	lootEqp.Modifiers = append(lootEqp.Modifiers, lootMod)
	loot.Equipment = append(loot.Equipment, lootEqp)
	lootNote := NewNote(loot, nil, false)
	loot.Notes = append(loot.Notes, lootNote)
	got = visits(loot)
	c.Equal(3, len(got), "a loot sheet's equipment, its modifier and its note are visited and nothing else")
	c.Equal(1, got[lootMod], "the loot equipment's modifier is visited once")
}

// unmatchedOwner is a data owner without a source matcher, as an equipment list file is.
type unmatchedOwner struct{}

func (unmatchedOwner) OwningEntity() *Entity      { return nil }
func (unmatchedOwner) SourceMatcher() *SrcMatcher { return nil }
func (unmatchedOwner) WeightUnit() fxp.WeightUnit { return fxp.Pound }

// useTestLibrary returns the root of the library under the key in the global set, adding one rooted in a temporary
// directory until the test ends if there is none.
func useTestLibrary(t *testing.T, key string) string {
	t.Helper()
	libs := GlobalSettings().Libraries
	if lib := libs.Lookup(key); lib != nil {
		return lib.Path()
	}
	lib := library.NewLibrary(key, "", "", key, t.TempDir())
	libs.Store(key, lib)
	t.Cleanup(func() { libs.Remove(key) })
	return lib.Path()
}

// stubLibrarySources makes the sources all that the matcher holds, as though loaded from the library file. The file is
// created, empty (see useTestLibrary), since a matcher keeps what it holds only while the file is there and unchanged.
func stubLibrarySources[T interface {
	Hashable
	ID() tid.TID
}](t *testing.T, sm *SrcMatcher, libFile LibraryFile, sources ...T) {
	t.Helper()
	p := filepath.Join(useTestLibrary(t, libFile.Library), filepath.FromSlash(libFile.Path))
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, nil, 0o640); err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	data := libSrcData{
		path:       p,
		timestamp:  stat.ModTime(),
		size:       stat.Size(),
		dataHashes: make(map[tid.TID]HashAndData, len(sources)),
	}
	for _, one := range sources {
		data.dataHashes[one.ID()] = HashAndData{Hash: Hash64(one), Data: one}
	}
	sm.libHashes = map[LibraryFile]libSrcData{libFile: data}
}

// isolateUnownedSrcMatcher empties the shared matcher for the test, putting back what it held when the test ends.
func isolateUnownedSrcMatcher(t *testing.T) {
	t.Helper()
	savedHashes, savedFiles := unownedSrcMatcher.libHashes, unownedSrcFiles
	t.Cleanup(func() { unownedSrcMatcher.libHashes, unownedSrcFiles = savedHashes, savedFiles })
	unownedSrcMatcher.libHashes, unownedSrcFiles = nil, nil
}

// registerTestFileTypes registers the extensions as file types until the test ends, standing in for the ux-layer
// registration a matcher relies on to tell what a library file holds.
func registerTestFileTypes(t *testing.T, extensions ...string) {
	t.Helper()
	savedRegistry := maps.Clone(fileTypeRegistry)
	savedKnown := KnownFileTypes
	t.Cleanup(func() {
		fileTypeRegistry = savedRegistry
		KnownFileTypes = savedKnown
	})
	for _, ext := range extensions {
		(&FileInfo{Name: "Test " + ext, UTI: &uti.DataType{Extensions: []string{ext}}, IsGCSData: true}).Register()
	}
}

// TestMatchWithoutSourceMatcher verifies that a nil source matcher finds nothing rather than crashing, and that
// MatchSource uses the matcher of the node's data owner, or the shared one when the node has no data owner or its data
// owner has no matcher.
func TestMatchWithoutSourceMatcher(t *testing.T) {
	c := check.New(t)
	isolateUnownedSrcMatcher(t)
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + EquipmentExt}
	var sm *SrcMatcher
	source := NewEquipment(nil, nil, false)
	source.Name = "Library"
	custom := NewEquipment(nil, nil, false)
	state, data := sm.Match(custom)
	c.Equal(srcstate.Custom, state, "a node without a source is custom, matcher or not")
	c.Nil(data)
	sourced := NewEquipment(nil, nil, false)
	sourced.Source = Source{LibraryFile: libFile, TID: source.TID}
	state, data = sm.Match(sourced)
	c.Equal(srcstate.Missing, state, "a nil matcher can't find the source")
	c.Nil(data)

	state, _ = MatchSource(sourced)
	c.Equal(srcstate.Missing, state, "a source no library holds can't be found")
	state, _ = MatchSource(custom)
	c.Equal(srcstate.Custom, state)
	sourced.Name = "Local"
	sourced.SyncWithSource()
	c.Equal("Local", sourced.Name, "syncing with nothing to match against must change nothing")

	stubLibrarySources(t, &unownedSrcMatcher, libFile, source)
	// A nil matcher must cope with a provider that has a library file to load.
	provider := NewEntity()
	provider.OtherEquipment = append(provider.OtherEquipment, sourced)
	sm.PrepareHashes(provider)

	sourced.SetDataOwner(unmatchedOwner{})
	state, data = MatchSource(sourced)
	c.Equal(srcstate.Mismatched, state, "the shared matcher is used for a data owner without one")
	c.Equal(any(source), data, "the library's copy comes back with the state")
	sourced.SetDataOwner(nil)
	state, _ = MatchSource(sourced)
	c.Equal(srcstate.Mismatched, state, "and for a node without a data owner")
	sourced.SyncWithSource()
	c.Equal("Library", sourced.Name, "which can then be synced")

	sourced.Name = "Local"
	e := NewEntity()
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)
	unownedSrcMatcher.libHashes = nil
	sourced.SetDataOwner(e)
	state, data = MatchSource(sourced)
	c.Equal(srcstate.Mismatched, state, "the data owner's matcher is used")
	c.Equal(any(source), data, "the library's copy comes back with the state")
}

// TestSourceHashesFollowTheLibraryFile verifies that a matcher loads the hashes of a library file when first asked
// about a node from it, keeps them while the file stays as it is, loads them again once the file has been changed or
// its library moved, and drops them once the file or its library is gone.
func TestSourceHashesFollowTheLibraryFile(t *testing.T) {
	c := check.New(t)
	isolateUnownedSrcMatcher(t)
	registerTestFileTypes(t, TraitsExt)
	libFile := LibraryFile{Library: "Test Library", Path: "Traits/Test" + TraitsExt}
	root := useTestLibrary(t, libFile.Library)
	// Each save sets the file's time explicitly, since two saves may fall within the file system's resolution.
	modTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	lib := NewTrait(nil, nil, false)
	lib.Name = "Claws"
	save := func(dir string) string {
		p := filepath.Join(dir, filepath.FromSlash(libFile.Path))
		c.NoError(os.MkdirAll(filepath.Dir(p), 0o750))
		c.NoError(SaveTraits([]*Trait{lib}, p))
		c.NoError(os.Chtimes(p, modTime, modTime))
		return p
	}
	p := save(root)
	local := lib.Clone(libFile, nil, nil, Reference)
	stateOf := func() srcstate.Value {
		state, _ := MatchSource(local)
		return state
	}
	// libName returns the name of the library's copy as matched, or "" if it wasn't found.
	libName := func() string {
		if _, data := MatchSource(local); data != nil {
			if trait, ok := data.(*Trait); ok {
				return trait.Name
			}
		}
		return ""
	}
	loaded := func() any { return unownedSrcMatcher.libHashes[libFile].dataHashes[lib.TID].Data }

	c.Equal(srcstate.Matched, stateOf(), "the file is loaded when a node from it is first matched")
	first := loaded()
	c.NotNil(first)
	c.Equal(srcstate.Matched, stateOf())
	c.True(first == loaded(), "a file that hasn't changed isn't loaded again")

	lib.Name = "Fangs"
	modTime = modTime.Add(time.Minute)
	save(root)
	c.Equal("Fangs", libName(), "a file saved since is loaded again, though it be the same size")
	c.Equal(srcstate.Mismatched, stateOf())
	lib.Name = "Talons"
	save(root)
	c.Equal("Talons", libName(), "as is one of another size, though its time be the same")

	// Once the library has moved, the library file is another file, though it be of the same size and time.
	lib.Name = "Spikes"
	moved := t.TempDir()
	save(moved)
	testLib := GlobalSettings().Libraries.Lookup(libFile.Library)
	c.NoError(testLib.SetPath(moved))
	c.Equal("Spikes", libName(), "the file of a library that has moved is loaded from where it is now")
	c.NoError(testLib.SetPath(root))
	c.Equal("Talons", libName())

	c.NoError(os.Remove(p))
	c.Equal(srcstate.Missing, stateOf(), "the source can't be found once its file is gone")
	_, held := unownedSrcMatcher.libHashes[libFile]
	c.False(held, "and what was loaded from the file is dropped")

	lib.Name = "Claws"
	save(root)
	c.Equal(srcstate.Matched, stateOf(), "the file is loaded again once it is back")
	GlobalSettings().Libraries.Remove(libFile.Library)
	c.Equal(srcstate.Missing, stateOf(), "the source can't be found once its library is gone")
	_, held = unownedSrcMatcher.libHashes[libFile]
	c.False(held, "and what was loaded from the library's file is dropped")

	// A data owner's matcher drops it too, when asked to prepare the hashes of what the data owner holds.
	save(useTestLibrary(t, libFile.Library))
	e := NewEntity()
	owned := lib.Clone(libFile, e, nil, Reference)
	e.Traits = append(e.Traits, owned)
	e.SourceMatcher().PrepareHashes(e)
	state, _ := e.SourceMatcher().Match(owned)
	c.Equal(srcstate.Matched, state, "precondition: the data owner's matcher has loaded the file")
	GlobalSettings().Libraries.Remove(libFile.Library)
	e.SourceMatcher().PrepareHashes(e)
	state, _ = e.SourceMatcher().Match(owned)
	c.Equal(srcstate.Missing, state, "a data owner's matcher drops the file of a library that is gone as well")
}

// TestSharedMatcherHoldsOnlyTheFilesLastUsed verifies that the shared matcher holds no more than maxUnownedSrcFiles
// library files, letting go of the one used longest ago.
func TestSharedMatcherHoldsOnlyTheFilesLastUsed(t *testing.T) {
	c := check.New(t)
	isolateUnownedSrcMatcher(t)
	registerTestFileTypes(t, NotesExt)
	root := useTestLibrary(t, "Test Library")
	notes := make([]*Note, maxUnownedSrcFiles+1)
	for i := range notes {
		lib := NewNote(nil, nil, false)
		libFile := LibraryFile{Library: "Test Library", Path: fmt.Sprintf("Notes %d%s", i, NotesExt)}
		c.NoError(SaveNotes([]*Note{lib}, filepath.Join(root, libFile.Path)))
		notes[i] = lib.Clone(libFile, nil, nil, Reference)
	}
	held := func(i int) bool {
		_, ok := unownedSrcMatcher.libHashes[notes[i].Source.LibraryFile]
		return ok
	}
	for i := range maxUnownedSrcFiles {
		state, _ := MatchSource(notes[i])
		c.Equal(srcstate.Matched, state)
	}
	c.Equal(maxUnownedSrcFiles, len(unownedSrcMatcher.libHashes))
	state, _ := MatchSource(notes[0])
	c.Equal(srcstate.Matched, state)
	state, _ = MatchSource(notes[maxUnownedSrcFiles])
	c.Equal(srcstate.Matched, state, "a file beyond those held is still matched")
	c.Equal(maxUnownedSrcFiles, len(unownedSrcMatcher.libHashes), "by letting go of another")
	c.True(held(0), "the file used again since is kept")
	c.False(held(1), "the one used longest ago is let go")
	state, _ = MatchSource(notes[1])
	c.Equal(srcstate.Matched, state, "and is loaded again when next needed")

	gone := NewNote(nil, nil, false)
	gone.Source = Source{Library: "Test Library", Path: "Gone" + NotesExt, TID: notes[0].TID}
	state, _ = MatchSource(gone)
	c.Equal(srcstate.Missing, state)
	c.Equal(maxUnownedSrcFiles, len(unownedSrcMatcher.libHashes), "a file that can't be found pushes out none that can")
	c.Equal(maxUnownedSrcFiles, len(unownedSrcFiles))
	c.True(held(3), "so the file used longest ago is still held")
}

// TestBatchSourceMatchesLoadsEachFileOnce verifies that a batch of matches loads the library file of each node without
// a source matcher of its own just once, however many files the nodes come from and in whatever order they are matched,
// where the shared matcher loads a file again each time the others have pushed it out. The batch uses what the shared
// matcher already holds, leaves it as it was, and the files are checked again once the batch is done.
func TestBatchSourceMatchesLoadsEachFileOnce(t *testing.T) {
	c := check.New(t)
	isolateUnownedSrcMatcher(t)
	registerTestFileTypes(t, NotesExt)
	root := useTestLibrary(t, "Test Library")
	notes := make([]*Note, maxUnownedSrcFiles+2)
	paths := make([]string, len(notes))
	for i := range notes {
		lib := NewNote(nil, nil, false)
		libFile := LibraryFile{Library: "Test Library", Path: fmt.Sprintf("Notes %d%s", i, NotesExt)}
		paths[i] = filepath.Join(root, libFile.Path)
		c.NoError(SaveNotes([]*Note{lib}, paths[i]))
		notes[i] = lib.Clone(libFile, nil, nil, Reference)
	}
	// libCopy returns the library's copy of the note as matched, which is another one each time its file is loaded.
	libCopy := func(i int) any {
		state, data := MatchSource(notes[i])
		c.Equal(srcstate.Matched, state)
		return data
	}
	last := len(notes) - 1
	first := libCopy(0)
	for i := 1; i <= last; i++ {
		libCopy(i)
	}
	c.True(first != libCopy(0), "precondition: outside of a batch, a file the others pushed out is loaded again")
	sharedCopy := libCopy(last)
	sharedFiles := slices.Clone(unownedSrcFiles)

	done := BatchSourceMatches()
	copies := make([]any, len(notes))
	for i := range notes {
		copies[i] = libCopy(i)
	}
	for i := range notes {
		c.True(copies[i] == libCopy(i), "each file is loaded just once in a batch, however many are used in between")
	}
	c.True(sharedCopy == copies[last], "a file the shared matcher holds isn't loaded again for the batch")
	c.NoError(os.Remove(paths[0]))
	c.True(copies[0] == libCopy(0), "a file is checked just once in a batch")
	gone := NewNote(nil, nil, false)
	gone.Source = Source{Library: "Test Library", Path: "Gone" + NotesExt, TID: notes[0].TID}
	state, _ := MatchSource(gone)
	c.Equal(srcstate.Missing, state, "a source whose file can't be found is still missing in a batch")
	BatchSourceMatches()()
	c.True(copies[0] == libCopy(0), "a batch begun within another doesn't end it")
	c.Equal(sharedFiles, unownedSrcFiles, "the batch leaves the shared matcher as it was")
	c.Equal(len(sharedFiles), len(unownedSrcMatcher.libHashes))

	done()
	state, _ = MatchSource(notes[0])
	c.Equal(srcstate.Missing, state, "the files are checked again once the batch is done")
	c.True(sharedCopy == libCopy(last), "and the shared matcher is used again")
}

// watchedTemplate is a template that calls onSourceMatcher whenever it is asked for its source matcher.
type watchedTemplate struct {
	*Template
	onSourceMatcher func()
}

func (w *watchedTemplate) DataOwner() DataOwner { return w }

func (w *watchedTemplate) SourceMatcher() *SrcMatcher {
	w.onSourceMatcher()
	return w.Template.SourceMatcher()
}

// TestSyncWithLibrarySourcesChecksEachFileOnce verifies that syncing every node a provider holds checks the library
// files they are sourced from just once, up front, rather than again for each node, by taking the file away as soon as
// it has been loaded, and that the file is checked again for a node matched once the sync is done.
func TestSyncWithLibrarySourcesChecksEachFileOnce(t *testing.T) {
	c := check.New(t)
	registerTestFileTypes(t, TraitsExt)
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + TraitsExt}
	p := filepath.Join(useTestLibrary(t, libFile.Library), libFile.Path)
	libTraits := []*Trait{NewTrait(nil, nil, false), NewTrait(nil, nil, false)}
	libTraits[0].Name = "Claws"
	libTraits[1].Name = "Fangs"
	c.NoError(SaveTraits(libTraits, p))
	tmpl := &watchedTemplate{Template: NewTemplate()}
	sm := tmpl.Template.SourceMatcher()
	// Each node asks its data owner for the matcher as it is synced, which is when a check of its file would be made.
	tmpl.onSourceMatcher = func() {
		if _, loaded := sm.libHashes[libFile]; loaded {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				t.Error(err)
			}
		}
	}
	for _, one := range libTraits {
		local := one.Clone(libFile, tmpl, nil, Reference)
		local.Name += " (old)"
		tmpl.Traits = append(tmpl.Traits, local)
	}

	syncWithLibrarySources(tmpl)
	_, err := os.Stat(p)
	c.True(os.IsNotExist(err), "precondition: the file was taken away during the sync")
	c.Equal("Claws", tmpl.Traits[0].Name, "a node is synced with the file as it was checked up front")
	c.Equal("Fangs", tmpl.Traits[1].Name, "as is every node after it")

	tmpl.Traits[0].Name = "Claws (old)"
	state, _ := MatchSource(tmpl.Traits[0])
	c.Equal(srcstate.Missing, state, "the file is checked again for a node matched once the sync is done")
}

// TestCloneWithoutParentHashesTheSame verifies that a copy of a node made without its parent, as an editor makes to
// compare its pending changes against the library, hashes the same as the node, for every kind of node with a source.
func TestCloneWithoutParentHashesTheSame(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	same := func(name string, original, clone Hashable) {
		c.Equal(Hash64(original), Hash64(clone), name)
	}
	weapons := func(owner WeaponOwner) []*Weapon {
		melee := NewWeapon(owner, true)
		melee.Usage = "Swung"
		melee.Defaults = []*SkillDefault{
			newSkillDefaultTo("Broadsword", "", true, -fxp.Two),
			{DefaultType: DexterityID, Name: textCriteria(criteria.IsText, "Leftover"), Modifier: -fxp.Five},
		}
		ranged := NewWeapon(owner, false)
		ranged.Usage = "Thrown"
		ranged.Defaults = []*SkillDefault{newSkillDefaultTo("Thrown Weapon", "Knife", false, 0)}
		return []*Weapon{melee, ranged}
	}
	features := func() Features {
		return Features{newSkillBonusTo("Alchemy", fxp.Two), NewAttributeBonus(StrengthID)}
	}
	prereqs := func() *PrereqList {
		list := NewPrereqList()
		either := NewPrereqList()
		either.All = false
		either.Parent = list
		for _, name := range []string{"Magery", "Power Investiture"} {
			one := NewTraitPrereq()
			one.NameCriteria.Qualifier = name
			one.Parent = either
			either.Prereqs = append(either.Prereqs, one)
		}
		skill := NewSkillPrereq()
		skill.NameCriteria.Qualifier = "Thaumatology"
		skill.Parent = list
		list.Prereqs = Prereqs{either, skill}
		return list
	}

	traitParent := NewTrait(e, nil, true)
	trait := NewTrait(e, traitParent, false)
	trait.Name = "Trait"
	trait.BasePoints = fxp.Ten
	trait.Tags = []string{"Physical"}
	trait.Weapons = weapons(trait)
	trait.Features = features()
	trait.Prereq = prereqs()
	trait.Modifiers = []*TraitModifier{NewTraitModifier(e, nil, false)}
	trait.Modifiers[0].Name = "Own Modifier"
	trait.Modifiers[0].Features = features()
	same("trait", trait, trait.Clone(LibraryFile{}, e, nil, Copy))
	traitParent.Name = "Trait Container"
	traitParent.Prereq = prereqs()
	traitParent.Children = []*Trait{trait}
	same("trait container", traitParent, traitParent.Clone(LibraryFile{}, e, nil, Copy))
	skillParent := NewSkill(e, nil, true)
	skill := NewSkill(e, skillParent, false)
	skill.Name = "Skill"
	skill.Tags = []string{"Combat"}
	skill.Defaults = []*SkillDefault{
		newSkillDefaultTo("Knife", "", true, -fxp.Three),
		{DefaultType: DexterityID, Name: textCriteria(criteria.IsText, "Leftover"), Modifier: -fxp.Five},
	}
	skill.Weapons = weapons(skill)
	skill.Features = features()
	skill.Prereq = prereqs()
	same("skill", skill, skill.Clone(LibraryFile{}, e, nil, Copy))
	technique := NewTechnique(e, skillParent, "Karate")
	technique.Name = "Technique"
	technique.Weapons = weapons(technique)
	technique.Features = features()
	technique.Prereq = prereqs()
	same("technique", technique, technique.Clone(LibraryFile{}, e, nil, Copy))
	spellParent := NewSpell(e, nil, true)
	spell := NewSpell(e, spellParent, false)
	spell.Name = "Spell"
	spell.College = []string{"Fire"}
	spell.Weapons = weapons(spell)
	spell.Features = features()
	spell.Prereq = prereqs()
	same("spell", spell, spell.Clone(LibraryFile{}, e, nil, Copy))
	eqpParent := NewEquipment(e, nil, true)
	eqp := NewEquipment(e, eqpParent, false)
	eqp.Name = "Equipment"
	eqp.BaseValue = "10"
	eqp.Tags = []string{"Weapon"}
	eqp.Weapons = weapons(eqp)
	eqp.Features = features()
	eqp.Prereq = prereqs()
	eqp.Modifiers = []*EquipmentModifier{NewEquipmentModifier(e, nil, false)}
	eqp.Modifiers[0].Name = "Own Modifier"
	eqp.Modifiers[0].Features = features()
	same("equipment", eqp, eqp.Clone(LibraryFile{}, e, nil, Copy))
	eqpParent.Name = "Equipment Container"
	eqpParent.Weapons = weapons(eqpParent)
	eqpParent.Features = features()
	eqpParent.Prereq = prereqs()
	eqpParent.Children = []*Equipment{eqp}
	same("equipment container", eqpParent, eqpParent.Clone(LibraryFile{}, e, nil, Copy))
	noteParent := NewNote(e, nil, true)
	note := NewNote(e, noteParent, false)
	note.MarkDown = "Note"
	same("note", note, note.Clone(LibraryFile{}, e, nil, Copy))
	traitModParent := NewTraitModifierChoice(e, nil)
	traitMod := NewTraitModifier(e, traitModParent, false)
	traitMod.Name = "Trait Modifier"
	traitMod.Tags = []string{"Enhancement"}
	traitMod.Features = features()
	same("trait modifier", traitMod, traitMod.Clone(LibraryFile{}, e, nil, Copy))
	traitModParent.Name = "Trait Modifier Choice"
	traitModParent.Children = []*TraitModifier{traitMod}
	same("trait modifier choice", traitModParent, traitModParent.Clone(LibraryFile{}, e, nil, Copy))
	eqpModParent := NewEquipmentModifierChoice(e, nil)
	eqpMod := NewEquipmentModifier(e, eqpModParent, false)
	eqpMod.Name = "Equipment Modifier"
	eqpMod.Tags = []string{"Quality"}
	eqpMod.Features = features()
	same("equipment modifier", eqpMod, eqpMod.Clone(LibraryFile{}, e, nil, Copy))
	eqpModParent.Name = "Equipment Modifier Choice"
	eqpModParent.Children = []*EquipmentModifier{eqpMod}
	same("equipment modifier choice", eqpModParent, eqpModParent.Clone(LibraryFile{}, e, nil, Copy))
}

// TestUnsavedWeaponsMatchOnceLoaded verifies that a node whose weapons have never been through a file hashes the same
// as it does once saved and loaded, so that a copy of it made before the save matches the library source read back from
// disk. The weapons are new ones of each kind, the natural attacks, and one edited to leave values behind in damage
// fields that aren't written out. Also verifies that a loaded weapon given fragmentation shows no armor divisor for it,
// and that the file is written as it always has been.
func TestUnsavedWeaponsMatchOnceLoaded(t *testing.T) {
	c := check.New(t)
	sk := NewSkill(nil, nil, false)
	sk.Name = "Stage Combat"
	edited := NewWeapon(sk, false)
	sk.Weapons = []*Weapon{NewWeapon(sk, true), NewWeapon(sk, false), newBite(sk), newPunch(sk), newKick(sk), edited}
	var edit Weapon
	edit.CopyFrom(edited)
	edit.Damage.ArmorDivisor = 0
	edit.Damage.FragmentationArmorDivisor = fxp.Two
	edit.Damage.FragmentationType = "cut"
	edit.ApplyTo(edited)

	p := filepath.Join(t.TempDir(), "Skills"+SkillsExt)
	c.NoError(SaveSkills([]*Skill{sk}, p))
	data, err := os.ReadFile(p)
	c.NoError(err)
	for _, key := range []string{"st_mul", "armor_divisor", "fragmentation"} {
		c.NotContains(string(data), key, "values that go without saying aren't written")
	}
	loaded, err := NewSkillsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p))
	c.NoError(err)
	c.Equal(1, len(loaded))
	c.Equal(len(sk.Weapons), len(loaded[0].Weapons))
	for i, w := range loaded[0].Weapons {
		c.Equal(sk.Weapons[i].Damage.WeaponDamageData, w.Damage.WeaponDamageData,
			"weapon %d: the damage is the same once loaded", i)
	}
	c.Equal(Hash64(sk), Hash64(loaded[0]), "the skill hashes the same once loaded")

	p2 := filepath.Join(t.TempDir(), "Skills"+SkillsExt)
	c.NoError(SaveSkills(loaded, p2))
	data2, err := os.ReadFile(p2)
	c.NoError(err)
	c.Equal(string(data), string(data2), "and is written back the same")

	e := NewEntity()
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + SkillsExt}
	local := sk.Clone(libFile, e, nil, Reference)
	local.Source = Source{LibraryFile: libFile, TID: loaded[0].TID}
	e.Skills = append(e.Skills, local)
	stubLibrarySources(t, e.SourceMatcher(), libFile, loaded[0])
	state, _ := MatchSource(local)
	c.Equal(srcstate.Matched, state, "a copy made before the save matches the source as loaded")

	edit = Weapon{}
	edit.CopyFrom(loaded[0].Weapons[0])
	c.Equal(fxp.One, edit.Damage.FragmentationArmorDivisor, "a loaded weapon has a fragmentation armor divisor of 1")
	edit.Damage.Fragmentation = "2d"
	c.Equal("thr cr [2d]", edit.Damage.String(), "so fragmentation added to it shows no armor divisor")
}

// TestLoadedTechniqueCopiesHashTheSame verifies that a technique loaded from a file hashes the same as copies of it,
// and that loading gives a name criteria only to a default that is skill-based, leaving a copy nothing to drop.
func TestLoadedTechniqueCopiesHashTheSame(t *testing.T) {
	c := check.New(t)
	attrBased := NewTechnique(nil, nil, "Karate")
	attrBased.Name = "Kicking"
	attrBased.TechniqueDefault = &SkillDefault{DefaultType: DexterityID, Modifier: -fxp.Two}
	skillBased := NewTechnique(nil, nil, "Karate")
	skillBased.Name = "Jump Kick"
	p := filepath.Join(t.TempDir(), "Techniques"+SkillsExt)
	c.NoError(SaveSkills([]*Skill{attrBased, skillBased}, p))
	loaded, err := NewSkillsFromFile(os.DirFS(filepath.Dir(p)), filepath.Base(p))
	c.NoError(err)
	c.Equal(2, len(loaded))
	c.True(loaded[0].TechniqueDefault.Name.IsZero(), "loading gives the attribute default no name criteria")
	c.Equal(criteria.IsText, loaded[1].TechniqueDefault.Name.Compare, "loading gives the skill default its name criteria")

	e := NewEntity()
	for _, one := range loaded {
		clone := one.Clone(LibraryFile{}, e, nil, Copy)
		c.Equal(Hash64(one), Hash64(clone), one.Name+": a copy hashes the same")
		var data SkillEditData
		data.CopyFrom(one)
		edited := one.Clone(LibraryFile{}, e, nil, Copy)
		data.ApplyTo(edited)
		c.Equal(Hash64(one), Hash64(edited), one.Name+": a copy passed through an editor's data hashes the same")
		synced := one.Clone(LibraryFile{}, e, nil, Copy)
		synced.Name = "Changed"
		synced.Source = Source{Library: "Test Library", Path: "Techniques" + SkillsExt, TID: one.TID}
		stubLibrarySources(t, e.SourceMatcher(), synced.Source.LibraryFile, one)
		synced.SyncWithSource()
		state, _ := MatchSource(synced)
		c.Equal(srcstate.Matched, state, one.Name+": syncing makes a copy match its source")
	}
}
