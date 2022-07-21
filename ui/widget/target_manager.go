package widget

import (
	"strconv"

	"github.com/richardwilkes/unison"
)

// TargetIDKey is the key that should be used in the client data of panels to identify targets.
const TargetIDKey = "target-id"

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

// Find searches the tree of panels starting at the root, looking for a specific ID.
func (t *TargetMgr) Find(id string) *unison.Panel {
	return t.find(t.root, id)
}

func (t *TargetMgr) find(p *unison.Panel, id string) *unison.Panel {
	if v, exists := p.ClientData()[TargetIDKey]; exists {
		if s, ok := v.(string); ok && id == s {
			return p
		}
	}
	for _, child := range p.Children() {
		if found := t.find(child, id); found != nil {
			return found
		}
	}
	return nil
}
