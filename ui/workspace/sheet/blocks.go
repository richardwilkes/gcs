package sheet

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/unison"
)

func createTopBlock(entity *gurps.Entity, targetMgr *widget.TargetMgr) (page *Page, modifiedFunc func()) {
	page = NewPage(entity)
	var top *unison.Panel
	top, modifiedFunc = createFirstRow(entity, targetMgr)
	page.AddChild(top)
	page.AddChild(createSecondRow(entity, targetMgr))
	return page, modifiedFunc
}

func createFirstRow(entity *gurps.Entity, targetMgr *widget.TargetMgr) (top *unison.Panel, modifiedFunc func()) {
	right := unison.NewPanel()
	right.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 1,
		VSpacing: 1,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})
	right.AddChild(NewIdentityPanel(entity, targetMgr))
	miscPanel := NewMiscPanel(entity, targetMgr)
	right.AddChild(miscPanel)
	right.AddChild(NewPointsPanel(entity, targetMgr))
	right.AddChild(NewDescriptionPanel(entity, targetMgr))

	top = unison.NewPanel()
	portraitPanel := NewPortraitPanel(entity)
	top.SetLayout(&portraitLayout{
		portrait: portraitPanel,
		rest:     right,
	})
	top.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})
	top.AddChild(portraitPanel)
	top.AddChild(right)

	return top, miscPanel.UpdateModified
}

func createSecondRow(entity *gurps.Entity, targetMgr *widget.TargetMgr) *unison.Panel {
	p := unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns:  4,
		HSpacing: 1,
		VSpacing: 1,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})

	endWrapper := unison.NewPanel()
	endWrapper.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	endWrapper.SetLayoutData(&unison.FlexLayoutData{
		VSpan:  3,
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
	})
	endWrapper.AddChild(NewEncumbrancePanel(entity))
	endWrapper.AddChild(NewLiftingPanel(entity))

	p.AddChild(NewPrimaryAttrPanel(entity, targetMgr))
	p.AddChild(NewSecondaryAttrPanel(entity, targetMgr))
	p.AddChild(NewBodyPanel(entity))
	p.AddChild(endWrapper)
	p.AddChild(NewDamagePanel(entity))
	p.AddChild(NewPointPoolsPanel(entity, targetMgr))

	return p
}
