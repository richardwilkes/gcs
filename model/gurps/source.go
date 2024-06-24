package gurps

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/tid"
)

var _ json.Omitter = &Source{}

// LibraryFile holds the library and path to a file.
type LibraryFile struct {
	Library string `json:"library"`
	Path    string `json:"path"`
}

// Source holds a reference to the source of a particular piece of data.
type Source struct {
	LibraryFile
	TID tid.TID `json:"id"`
}

// ShouldOmit implements json.Omitter.
func (s Source) ShouldOmit() bool {
	return s.TID == "" || s.Library == "" || s.Path == ""
}

func (s Source) collectInto(m map[LibraryFile][]tid.TID) {
	if !s.ShouldOmit() {
		m[s.LibraryFile] = append(m[s.LibraryFile], s.TID)
	}
}

// CollectSourceHashes returns a map of TIDs to hashes for all sources found in the ListProvider.
func CollectSourceHashes(provider ListProvider) map[tid.TID]uint64 {
	m := make(map[LibraryFile][]tid.TID)
	Traverse(func(t *Trait) bool {
		t.Source.collectInto(m)
		Traverse(func(mod *TraitModifier) bool {
			mod.Source.collectInto(m)
			return false
		}, false, false, t.Modifiers...)
		return false
	}, false, false, provider.TraitList()...)
	Traverse(func(s *Skill) bool {
		s.Source.collectInto(m)
		return false
	}, false, false, provider.SkillList()...)
	Traverse(func(s *Spell) bool {
		s.Source.collectInto(m)
		return false
	}, false, false, provider.SpellList()...)
	Traverse(func(e *Equipment) bool {
		e.Source.collectInto(m)
		Traverse(func(mod *EquipmentModifier) bool {
			mod.Source.collectInto(m)
			return false
		}, false, false, e.Modifiers...)
		return false
	}, false, false, provider.CarriedEquipmentList()...)
	Traverse(func(e *Equipment) bool {
		e.Source.collectInto(m)
		Traverse(func(mod *EquipmentModifier) bool {
			mod.Source.collectInto(m)
			return false
		}, false, false, e.Modifiers...)
		return false
	}, false, false, provider.OtherEquipmentList()...)
	Traverse(func(n *Note) bool {
		n.Source.collectInto(m)
		return false
	}, false, false, provider.NoteList()...)
	libs := GlobalSettings().Libraries()
	result := make(map[tid.TID]uint64)
	for libFile, tids := range m {
		if lib, ok := libs[libFile.Library]; ok {
			need := make(map[tid.TID]bool)
			for _, id := range tids {
				need[id] = true
			}
			p := filepath.Join(lib.Path(), libFile.Path)
			dir := os.DirFS(filepath.Dir(p))
			file := filepath.Base(p)
			fi := FileInfoFor(p)
			switch fi.Extensions[0] {
			case TraitsExt:
				if data, err := NewTraitsFromFile(dir, file); err == nil {
					NodesToHashesByID(need, result, data...)
					Traverse(func(t *Trait) bool {
						NodesToHashesByID(need, result, t.Modifiers...)
						return false
					}, false, false, data...)
				}
			case TraitModifiersExt:
				if data, err := NewTraitModifiersFromFile(dir, file); err == nil {
					NodesToHashesByID(need, result, data...)
				}
			case SkillsExt:
				if data, err := NewSkillsFromFile(dir, file); err == nil {
					NodesToHashesByID(need, result, data...)
				}
			case SpellsExt:
				if data, err := NewSpellsFromFile(dir, file); err == nil {
					NodesToHashesByID(need, result, data...)
				}
			case EquipmentExt:
				if data, err := NewEquipmentFromFile(dir, file); err == nil {
					NodesToHashesByID(need, result, data...)
					Traverse(func(e *Equipment) bool {
						NodesToHashesByID(need, result, e.Modifiers...)
						return false
					}, false, false, data...)
				}
			case EquipmentModifiersExt:
				if data, err := NewEquipmentModifiersFromFile(dir, file); err == nil {
					NodesToHashesByID(need, result, data...)
				}
			case NotesExt:
				if data, err := NewNotesFromFile(dir, file); err == nil {
					NodesToHashesByID(need, result, data...)
				}
			}
		}
	}
	return result
}

/*

States:

- Custom
- Sourced, matching
- Sourced, mismatched

*/
