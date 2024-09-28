package gurps

import (
	"os"
	"path/filepath"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
)

var _ json.Omitter = &Source{}

// LibraryFile holds the library and path to a file.
type LibraryFile struct {
	Library string `json:"library"`
	Path    string `json:"path"`
}

// SourcedID holds a TID and an optional Source.
type SourcedID struct {
	TID    tid.TID `json:"id"`
	Source Source  `json:"source,omitempty"`
}

// Source holds a reference to the source of a particular piece of data.
type Source struct {
	LibraryFile
	TID tid.TID `json:"id"`
}

type libSrcData struct {
	timestamp  time.Time
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

// ShouldOmit implements json.Omitter.
func (s Source) ShouldOmit() bool {
	return s.TID == "" || s.Library == "" || s.Path == ""
}

func (s Source) collectInto(m map[LibraryFile]struct{}) {
	if !s.ShouldOmit() {
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

// PrepareHashes for the given ListProvider.
func (sm *SrcMatcher) PrepareHashes(provider ListProvider) {
	neededLibs := make(map[LibraryFile]struct{})
	Traverse(func(t *Trait) bool {
		t.Source.collectInto(neededLibs)
		Traverse(func(mod *TraitModifier) bool {
			mod.Source.collectInto(neededLibs)
			return false
		}, false, false, t.Modifiers...)
		return false
	}, false, false, provider.TraitList()...)
	Traverse(func(s *Skill) bool {
		s.Source.collectInto(neededLibs)
		return false
	}, false, false, provider.SkillList()...)
	Traverse(func(s *Spell) bool {
		s.Source.collectInto(neededLibs)
		return false
	}, false, false, provider.SpellList()...)
	Traverse(func(e *Equipment) bool {
		e.Source.collectInto(neededLibs)
		Traverse(func(mod *EquipmentModifier) bool {
			mod.Source.collectInto(neededLibs)
			return false
		}, false, false, e.Modifiers...)
		return false
	}, false, false, provider.CarriedEquipmentList()...)
	Traverse(func(e *Equipment) bool {
		e.Source.collectInto(neededLibs)
		Traverse(func(mod *EquipmentModifier) bool {
			mod.Source.collectInto(neededLibs)
			return false
		}, false, false, e.Modifiers...)
		return false
	}, false, false, provider.OtherEquipmentList()...)
	Traverse(func(n *Note) bool {
		n.Source.collectInto(neededLibs)
		return false
	}, false, false, provider.NoteList()...)
	libs := GlobalSettings().Libraries()
	if sm.libHashes == nil {
		sm.libHashes = make(map[LibraryFile]libSrcData)
	}
	for libFile := range neededLibs {
		lib, ok := libs[libFile.Library]
		if !ok {
			continue
		}
		p := filepath.Join(lib.Path(), libFile.Path)
		stat, err := os.Stat(p)
		if err != nil {
			delete(sm.libHashes, libFile)
			continue
		}
		modTime := stat.ModTime()
		if data, exists := sm.libHashes[libFile]; exists {
			if modTime.Compare(data.timestamp) >= 0 {
				continue // We've already loaded this file and it hasn't changed.
			}
			delete(sm.libHashes, libFile)
		}
		var srcData libSrcData
		srcData.timestamp = modTime
		srcData.dataHashes = make(map[tid.TID]HashAndData)
		dir := os.DirFS(filepath.Dir(p))
		file := filepath.Base(p)
		fi := FileInfoFor(p)
		if fi == nil || len(fi.Extensions) == 0 {
			continue
		}
		switch fi.Extensions[0] {
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
}

// Match returns the source state of the given data.
func (sm *SrcMatcher) Match(data SrcProvider) (state srcstate.Value, match any) {
	src := data.GetSource()
	if src.ShouldOmit() {
		return srcstate.Custom, nil
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

// AdjustSource adjusts the source of a SourcedID to match the given LibraryFile.
func (s *SourcedID) AdjustSource(from LibraryFile, original SourcedID, preserve bool) {
	if preserve {
		s.TID = original.TID
	}
	s.Source = original.Source
	if s.Source.Library == "" {
		s.Source.LibraryFile = from
		s.Source.TID = original.TID
	}
}
