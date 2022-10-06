package attrdef

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

type poolPanel struct {
	unison.Panel
	dockable *attributesDockable
	def      *gurps.AttributeDef
}

func newPoolPanel(dockable *attributesDockable, def *gurps.AttributeDef) *poolPanel {
	p := &poolPanel{
		dockable: dockable,
		def:      def,
	}
	p.Self = p
	p.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	p.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	for _, threshold := range def.Thresholds {
		p.AddChild(newThresholdPanel(p, threshold))
	}
	return p
}

func (p *poolPanel) addThreshold() {
	undo := &unison.UndoEdit[[]*gurps.PoolThreshold]{
		ID:       unison.NextUndoID(),
		EditName: i18n.Text("Add Pool Threshold"),
		UndoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.BeforeData)
		},
		RedoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.AfterData)
		},
		AbsorbFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold], other unison.Undoable) bool { return false },
	}
	undo.BeforeData = clonePoolThresholds(p.def.Thresholds)
	threshold := &gurps.PoolThreshold{KeyPrefix: p.dockable.targetMgr.NextPrefix()}
	p.def.Thresholds = append(p.def.Thresholds, threshold)
	newThreshold := newThresholdPanel(p, threshold)
	p.AddChild(newThreshold)
	if children := p.Children(); len(children) == 2 {
		children[0].Self.(*thresholdPanel).deleteButton.SetEnabled(true)
	}
	undo.AfterData = clonePoolThresholds(p.def.Thresholds)
	p.dockable.UndoManager().Add(undo)
	p.dockable.MarkModified(nil)
	p.dockable.MarkForLayoutAndRedraw()
	p.dockable.ValidateLayout()
	focus := newThreshold.Children()[2]
	focus.RequestFocus()
	focus.ScrollIntoView()
}

func (p *poolPanel) deleteThreshold(target *thresholdPanel) {
	i := p.IndexOfChild(target)
	target.RemoveFromParent()
	if children := p.Children(); len(children) == 1 {
		children[0].Self.(*thresholdPanel).deleteButton.SetEnabled(false)
	}
	undo := &unison.UndoEdit[[]*gurps.PoolThreshold]{
		ID:       unison.NextUndoID(),
		EditName: i18n.Text("Delete Pool Threshold"),
		UndoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.BeforeData)
		},
		RedoFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold]) {
			p.applyThresholds(e.AfterData)
		},
		AbsorbFunc: func(e *unison.UndoEdit[[]*gurps.PoolThreshold], other unison.Undoable) bool { return false },
	}
	undo.BeforeData = clonePoolThresholds(p.def.Thresholds)
	p.def.Thresholds = slices.Delete(p.def.Thresholds, i, i+1)
	undo.AfterData = clonePoolThresholds(p.def.Thresholds)
	p.dockable.UndoManager().Add(undo)
	p.dockable.MarkModified(nil)
}

func (p *poolPanel) applyThresholds(thresholds []*gurps.PoolThreshold) {
	p.def.Thresholds = clonePoolThresholds(thresholds)
	p.dockable.sync()
}

func clonePoolThresholds(in []*gurps.PoolThreshold) []*gurps.PoolThreshold {
	thresholds := make([]*gurps.PoolThreshold, len(in))
	for i, one := range in {
		thresholds[i] = one.Clone()
	}
	return thresholds
}
