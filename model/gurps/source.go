package gurps

import (
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
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

// SrcMatcher provides Source matching for a given ListProvider.
type SrcMatcher struct {
	hashes map[tid.TID]uint64
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

// NewSrcMatch returns a new SrcMatch for the given ListProvider.
func NewSrcMatch(provider ListProvider) *SrcMatcher {
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
	var s SrcMatcher
	s.hashes = make(map[tid.TID]uint64)
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
					NodesToHashesByID(need, s.hashes, data...)
					Traverse(func(t *Trait) bool {
						NodesToHashesByID(need, s.hashes, t.Modifiers...)
						return false
					}, false, false, data...)
				}
			case TraitModifiersExt:
				if data, err := NewTraitModifiersFromFile(dir, file); err == nil {
					NodesToHashesByID(need, s.hashes, data...)
				}
			case SkillsExt:
				if data, err := NewSkillsFromFile(dir, file); err == nil {
					NodesToHashesByID(need, s.hashes, data...)
				}
			case SpellsExt:
				if data, err := NewSpellsFromFile(dir, file); err == nil {
					NodesToHashesByID(need, s.hashes, data...)
				}
			case EquipmentExt:
				if data, err := NewEquipmentFromFile(dir, file); err == nil {
					NodesToHashesByID(need, s.hashes, data...)
					Traverse(func(e *Equipment) bool {
						NodesToHashesByID(need, s.hashes, e.Modifiers...)
						return false
					}, false, false, data...)
				}
			case EquipmentModifiersExt:
				if data, err := NewEquipmentModifiersFromFile(dir, file); err == nil {
					NodesToHashesByID(need, s.hashes, data...)
				}
			case NotesExt:
				if data, err := NewNotesFromFile(dir, file); err == nil {
					NodesToHashesByID(need, s.hashes, data...)
				}
			}
		}
	}
	return &s
}

// Match returns the source state of the given data.
func (s *SrcMatcher) Match(data HashableIDer) srcstate.Value {
	if h, ok := s.hashes[data.ID()]; ok {
		if h == Hash64(data) {
			return srcstate.Matched
		}
		return srcstate.Mismatched
	}
	return srcstate.Custom
}
