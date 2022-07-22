package widget

import (
	"strconv"

	"github.com/richardwilkes/unison"
)

// TargetMgr provides management of target panels.
type TargetMgr struct {
	root   *unison.Panel
	lastID int
}

// NewTargetMgr creates a new TargetMgr with the given root.
func NewTargetMgr(root unison.Paneler) *TargetMgr {
	return &TargetMgr{root: root.AsPanel()}
}

// NextPrefix returns the next unique prefix to use.
func (t *TargetMgr) NextPrefix() string {
	t.lastID++
	return strconv.Itoa(t.lastID) + ":"
}

// Find searches the tree of panels starting at the root, looking for a specific refKey.
func (t *TargetMgr) Find(refKey string) *unison.Panel {
	return t.root.FindRefKey(refKey)
}
