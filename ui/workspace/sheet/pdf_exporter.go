package sheet

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

var _ widget.Rebuildable = &pdfExporter{}

const pdfKey = "pdfKey"

type pdfExporter struct {
	unison.Panel
	entity      *gurps.Entity
	targetMgr   *widget.TargetMgr
	pages       []*Page
	currentPage int
}

func newPDFExporter(entity *gurps.Entity) *pdfExporter {
	p := &pdfExporter{entity: entity}
	p.targetMgr = widget.NewTargetMgr(p)
	pageSize := p.PageSize()
	r := unison.Rect{Size: pageSize}
	page, _ := createTopBlock(entity, p.targetMgr)
	p.AddChild(page)
	p.pages = append(p.pages, page)
	for _, col := range entity.SheetSettings.BlockLayout.ByRow() {
		startAt := make(map[string]int)
		for {
			rowPanel := unison.NewPanel()
			rowPanel.SetLayoutData(&unison.FlexLayoutData{
				HAlign: unison.FillAlignment,
				HGrab:  true,
			})
			for _, c := range col {
				switch c {
				case gurps.BlockLayoutReactionsKey:
					addRowPanel(rowPanel, NewReactionsPageList(entity), gurps.BlockLayoutReactionsKey, startAt)
				case gurps.BlockLayoutConditionalModifiersKey:
					addRowPanel(rowPanel, NewConditionalModifiersPageList(entity), gurps.BlockLayoutConditionalModifiersKey, startAt)
				case gurps.BlockLayoutMeleeKey:
					addRowPanel(rowPanel, NewMeleeWeaponsPageList(entity), gurps.BlockLayoutMeleeKey, startAt)
				case gurps.BlockLayoutRangedKey:
					addRowPanel(rowPanel, NewRangedWeaponsPageList(entity), gurps.BlockLayoutRangedKey, startAt)
				case gurps.BlockLayoutTraitsKey:
					addRowPanel(rowPanel, NewTraitsPageList(p, entity), gurps.BlockLayoutTraitsKey, startAt)
				case gurps.BlockLayoutSkillsKey:
					addRowPanel(rowPanel, NewSkillsPageList(p, entity), gurps.BlockLayoutSkillsKey, startAt)
				case gurps.BlockLayoutSpellsKey:
					addRowPanel(rowPanel, NewSpellsPageList(p, entity), gurps.BlockLayoutSpellsKey, startAt)
				case gurps.BlockLayoutEquipmentKey:
					addRowPanel(rowPanel, NewCarriedEquipmentPageList(p, entity), gurps.BlockLayoutEquipmentKey, startAt)
				case gurps.BlockLayoutOtherEquipmentKey:
					addRowPanel(rowPanel, NewOtherEquipmentPageList(p, entity), gurps.BlockLayoutOtherEquipmentKey, startAt)
				case gurps.BlockLayoutNotesKey:
					addRowPanel(rowPanel, NewNotesPageList(p, entity), gurps.BlockLayoutNotesKey, startAt)
				}
			}
			children := rowPanel.Children()
			if len(children) == 0 {
				break
			}
			rowPanel.SetLayout(&unison.FlexLayout{
				Columns:      len(children),
				HSpacing:     1,
				HAlign:       unison.FillAlignment,
				EqualColumns: true,
			})
			page.AddChild(rowPanel)
			page.SetFrameRect(r)
			page.MarkForLayoutRecursively()
			page.ValidateLayout()
			_, pref, _ := page.Sizes(unison.Size{Width: r.Width})
			excess := pref.Height - pageSize.Height
			if excess <= 0 {
				break // Not extending off the page, so move to the next row
			}
			remaining := (pageSize.Height - page.insets().Bottom) - rowPanel.FrameRect().Y
			startNewPage := false
			data := make([]*pdfState, len(children))
			for i, child := range children {
				data[i] = newPDFState(child)
				if remaining < data[i].minimum {
					startNewPage = true
				}
			}
			if startNewPage {
				// At least one of the columns can't fit at least the header plus the next row, so start a new page
				page.RemoveChild(rowPanel)
				page = NewPage(entity)
				p.AddChild(page)
				p.pages = append(p.pages, page)
				page.AddChild(rowPanel)
				page.SetFrameRect(r)
				page.MarkForLayoutRecursively()
				page.ValidateLayout()
				_, pref, _ = page.Sizes(unison.Size{Width: r.Width})
				if excess = pref.Height - pageSize.Height; excess <= 0 {
					break // Not extending off the page, so move to the next row
				}
				remaining = (pageSize.Height - page.insets().Bottom) - rowPanel.FrameRect().Y
			}
			for _, one := range data {
				allowed := remaining - one.overhead
				start, endBefore := one.helper.CurrentDrawRowRange()
				startAt[one.key()] = len(one.heights) // Assume all remaining fit
				for i := start; i < endBefore; i++ {
					allowed -= one.heights[i] + 1
					if allowed < 0 {
						// No more fit, mark it
						one.helper.SetDrawRowRange(start, xmath.Max(i, start+1))
						if i == start {
							// I have to guard against the case where a single row is so large it can't fit on a single
							// page on its own. In this case, I just let it flow off the end and drop that extra
							// content.
							//
							// TODO: In the future, see if I can do sub-row partitioning.
							i = start + 1
						}
						startAt[one.key()] = i
						break
					}
				}
			}
		}
	}
	for _, page = range p.pages {
		page.Force = true
		page.SetFrameRect(r)
		page.MarkForLayoutRecursively()
		page.ValidateLayout()
	}
	return p
}

type pdfHelper interface {
	OverheadHeight() float32
	RowHeights() []float32
	CurrentDrawRowRange() (start, endBefore int)
	SetDrawRowRange(start, endBefore int)
}

type pdfState struct {
	child    *unison.Panel
	helper   pdfHelper
	current  float32
	overhead float32
	minimum  float32
	heights  []float32
}

func newPDFState(child unison.Paneler) *pdfState {
	panel := child.AsPanel()
	helper := panel.Self.(pdfHelper) //nolint:errcheck // The only things used with this are pdfHelper-compliant
	state := &pdfState{
		child:    panel,
		helper:   helper,
		current:  panel.FrameRect().Height,
		overhead: helper.OverheadHeight(),
		heights:  helper.RowHeights(),
	}
	state.minimum = state.overhead
	start, _ := state.helper.CurrentDrawRowRange()
	if len(state.heights) > start {
		state.minimum += state.heights[start] + 1
	}
	return state
}

func (s *pdfState) key() string {
	return s.child.ClientData()[pdfKey].(string)
}

func addRowPanel[T gurps.NodeTypes](rowPanel *unison.Panel, list *PageList[T], key string, startAtMap map[string]int) {
	list.ClientData()[pdfKey] = key
	count := list.RowCount()
	startAt := startAtMap[key]
	if count > startAt {
		list.SetDrawRowRange(startAt, count)
		rowPanel.AddChild(list)
	}
}

func (p *pdfExporter) exportAsBytes() ([]byte, error) {
	stream := unison.NewMemoryStream()
	defer stream.Close()
	if err := p.export(stream); err != nil {
		return nil, err
	}
	return stream.Bytes(), nil
}

func (p *pdfExporter) exportAsFile(filePath string) error {
	stream, err := unison.NewFileStream(filePath)
	if err != nil {
		return err
	}
	defer stream.Close()
	return p.export(stream)
}

func (p *pdfExporter) export(stream unison.Stream) error {
	savedColorMode := unison.CurrentColorMode()
	unison.SetColorMode(unison.LightColorMode)
	unison.ThemeChanged()
	unison.RebuildDynamicColors()
	defer func() {
		unison.SetColorMode(savedColorMode)
		unison.ThemeChanged()
		unison.RebuildDynamicColors()
	}()
	if err := unison.CreatePDF(stream, &unison.PDFMetaData{
		Title:           p.entity.Profile.Name,
		Author:          toolbox.CurrentUserName(),
		Subject:         p.entity.Profile.Name,
		Keywords:        "GCS Character Sheet",
		Creator:         "GCS",
		RasterDPI:       300,
		EncodingQuality: 101,
	}, p); err != nil {
		return err
	}
	return nil
}

// HasPage implements unison.PageProvider.
func (p *pdfExporter) HasPage(pageNumber int) bool {
	p.currentPage = pageNumber
	return pageNumber > 0 && pageNumber <= len(p.pages)
}

// PageSize implements unison.PageProvider.
func (p *pdfExporter) PageSize() unison.Size {
	w, h := p.entity.SheetSettings.Page.Orientation.Dimensions(p.entity.SheetSettings.Page.Size.Dimensions())
	return unison.NewSize(w.Pixels(), h.Pixels())
}

// DrawPage implements unison.PageProvider.
func (p *pdfExporter) DrawPage(canvas *unison.Canvas, pageNumber int) error {
	p.currentPage = pageNumber
	if pageNumber > 0 && pageNumber <= len(p.pages) {
		page := p.pages[pageNumber-1]
		page.Draw(canvas, page.ContentRect(true))
		return nil
	}
	return errs.New("invalid page number")
}

func (p *pdfExporter) String() string {
	return ""
}

func (p *pdfExporter) Rebuild(_ bool) {
}
