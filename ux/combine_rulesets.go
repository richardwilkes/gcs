// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

// combineTreeSize is the size the two trees ask for. Without it they ask for whatever their contents need, which
// makes the dialog as tall as the user's libraries.
var combineTreeSize = geom.Size{Width: 340, Height: 300}

// combineNodeRole distinguishes what a row in one of the two trees represents.
type combineNodeRole int

const (
	combineRoleLibrary combineNodeRole = iota
	combineRoleSourceFolder
	combineRoleSourceFile
	combineRoleOutFolder
	combineRoleOutFile
	combineRoleComponent
)

var _ unison.TableRowData[*combineNode] = &combineNode{}

// combineNode is a row in either of the combine dialog's trees. The left tree holds libraries, their folders and
// their combinable data files; the right tree holds the staged output folders, files, and the components each file
// is built from.
type combineNode struct {
	id       tid.TID
	key      string // stable across rebuilds, for open-state and selection preservation
	title    string
	tooltip  string
	icon     string // an extension understood by gurps.FileInfoFor, or empty for a folder
	role     combineNodeRole
	parent   *combineNode
	children []*combineNode
	openMap  map[string]bool // shared per dialog; explicit open states, keyed by node key

	// Left-side payload: the library and the path relative to its root (slash form, "" for the library itself).
	lib     *gurps.Library
	relPath string

	// Right-side payload.
	outFile   *gurps.CombinedFile
	compIndex int
	comp      gurps.CombinedComponent

	// sizer, when set, re-fits the owning table's column after this node's open state changes, so that names
	// revealed by disclosure are not cut off at the width the table had when they were hidden.
	sizer func()
}

func newCombineNode(role combineNodeRole, key, title string, openMap map[string]bool, parent *combineNode) *combineNode {
	var kind byte
	switch role {
	case combineRoleLibrary:
		kind = kinds.NavigatorLibrary
	case combineRoleSourceFolder, combineRoleOutFolder, combineRoleOutFile:
		kind = kinds.NavigatorDirectory
	default:
		kind = kinds.NavigatorFile
	}
	n := &combineNode{
		id:      tid.MustNewTID(kind),
		key:     key,
		title:   title,
		role:    role,
		parent:  parent,
		openMap: openMap,
	}
	if parent != nil {
		parent.children = append(parent.children, n)
	}
	return n
}

// CloneForTarget implements unison.TableRowData. Dragging from the source tree copies: the clone carries the payload
// that identifies what was dragged, while the original stays where it was. Children are not cloned; anything that
// needs them is reconstructed from the plan or from disk when the staging is rebuilt after the drop.
func (n *combineNode) CloneForTarget(_ unison.Paneler, newParent *combineNode) *combineNode {
	clone := *n
	clone.id = tid.MustNewTID(n.id[0])
	clone.parent = newParent
	clone.children = nil
	clone.sizer = nil
	return &clone
}

// ID implements unison.TableRowData.
func (n *combineNode) ID() tid.TID {
	return n.id
}

// Parent implements unison.TableRowData.
func (n *combineNode) Parent() *combineNode {
	return n.parent
}

// SetParent implements unison.TableRowData.
func (n *combineNode) SetParent(parent *combineNode) {
	n.parent = parent
}

// CanHaveChildren implements unison.TableRowData.
func (n *combineNode) CanHaveChildren() bool {
	switch n.role {
	case combineRoleSourceFile, combineRoleComponent:
		return false
	default:
		return true
	}
}

// Children implements unison.TableRowData.
func (n *combineNode) Children() []*combineNode {
	return n.children
}

// SetChildren implements unison.TableRowData.
func (n *combineNode) SetChildren(children []*combineNode) {
	n.children = children
}

// CellDataForSort implements unison.TableRowData.
func (n *combineNode) CellDataForSort(col int) string {
	if col != 0 {
		return ""
	}
	return n.title
}

// ColumnCell implements unison.TableRowData.
func (n *combineNode) ColumnCell(_, col int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	if col != 0 {
		return unison.NewLabel()
	}
	ext := n.icon
	if ext == "" {
		if n.IsOpen() {
			ext = gurps.OpenFolder
		} else {
			ext = gurps.ClosedFolder
		}
	}
	size := unison.LabelFont.Size() + 5
	label := unison.NewLabel()
	label.OnBackgroundInk = foreground
	label.SetTitle(n.title)
	label.Drawable = &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(ext).SVG,
		Size: geom.NewSize(size, size),
	}
	if n.tooltip != "" {
		label.Tooltip = newWrappedTooltip(n.tooltip)
	}
	return label
}

// IsOpen implements unison.TableRowData. Nodes without an explicit state fall back to a per-role default: libraries
// and everything on the staging side start open, while source folders start closed, the way the navigator presents
// them.
func (n *combineNode) IsOpen() bool {
	if open, ok := n.openMap[n.key]; ok {
		return open
	}
	return n.role != combineRoleSourceFolder
}

// SetOpen implements unison.TableRowData.
func (n *combineNode) SetOpen(open bool) {
	n.openMap[n.key] = open
	if n.sizer != nil {
		n.sizer()
	}
}

// setCombineTreeSizer points every node of the subtree at the function that re-fits its table's column.
func setCombineTreeSizer(nodes []*combineNode, sizer func()) {
	for _, n := range nodes {
		n.sizer = sizer
		setCombineTreeSizer(n.children, sizer)
	}
}

// componentKey identifies a staged component across tree rebuilds.
func componentKey(comp gurps.CombinedComponent) string {
	return "C:" + comp.Library.Key() + ":" + strings.ToLower(comp.Path)
}

// topFolderOf splits a library-relative path into its source folder and the remainder.
func topFolderOf(relPath string) (folder, rest string) {
	folder, rest, _ = strings.Cut(relPath, "/")
	return folder, rest
}

// combineRulesetsDialog holds the state of one Create Combined Ruleset dialog.
type combineRulesetsDialog struct {
	plan               *gurps.CombinedPlan
	nameField          *unison.Field
	filterField        *unison.Field
	subfoldersCheckBox *unison.CheckBox
	left               *unison.Table[*combineNode]
	right              *unison.Table[*combineNode]
	addButton          *unison.Button
	removeButton       *unison.Button
	newFileButton      *unison.Button
	renameButton       *unison.Button
	moveButton         *unison.Button
	raiseButton        *unison.Button
	lowerButton        *unison.Button
	okButton           *unison.Button
	openMap            map[string]bool
	adjustLeftPending  bool
	adjustRightPending bool
}

// combineRulesets puts up the dialog for combining library data files into a new folder within the user library, then
// performs the combination. The left tree offers every combinable data file of every library; the right tree stages
// the output: which files will be written, and which components each is built from.
func combineRulesets() {
	d := &combineRulesetsDialog{
		plan:    &gurps.CombinedPlan{},
		openMap: make(map[string]bool),
	}
	d.run()
}

// run constructs and runs the modal dialog.
func (d *combineRulesetsDialog) run() {
	d.nameField = unison.NewField()
	d.nameField.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.Size{Width: 300},
		HAlign:  align.Fill,
		HGrab:   true,
	})

	d.subfoldersCheckBox = unison.NewCheckBox()
	d.subfoldersCheckBox.SetTitle(i18n.Text("Include files found in subfolders when adding a folder"))

	d.left = unison.NewTable(&unison.SimpleTableModel[*combineNode]{})
	d.left.PreventUserColumnResize = true
	d.left.ShowFirstColumnDivider = false
	d.left.ShowLastColumnDivider = false
	d.left.Columns = make([]unison.ColumnInfo, 1)
	d.right = unison.NewTable(&unison.SimpleTableModel[*combineNode]{})
	d.right.PreventUserColumnResize = true
	d.right.ShowFirstColumnDivider = false
	d.right.ShowLastColumnDivider = false
	d.right.Columns = make([]unison.ColumnInfo, 1)

	d.filterField = NewSearchField(i18n.Text("Filter"), func(_, _ *unison.FieldState) { d.refreshLeft() })
	d.refreshLeft()
	if len(d.left.RootRows()) == 0 {
		unison.WarningDialogWithMessage(i18n.Text("No library data files found"),
			i18n.Text("There are no data files in your libraries to combine."))
		return
	}

	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})

	intro := unison.NewLabel()
	intro.SetTitle(i18n.Text("Stage library data files on the right, then combine them into a new folder in the User Library."))
	panel.AddChild(intro)

	nameRow := unison.NewPanel()
	nameRow.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	nameRow.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	nameLabel := unison.NewLabel()
	nameLabel.SetTitle(i18n.Text("Name:"))
	nameRow.AddChild(nameLabel)
	nameRow.AddChild(d.nameField)
	panel.AddChild(nameRow)

	lists := unison.NewPanel()
	lists.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	lists.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	panel.AddChild(lists)

	lists.AddChild(newCombineRulesetsListHeader(i18n.Text("Available Rulesets")))
	lists.AddChild(unison.NewPanel()) // Spacer above the add/remove buttons
	lists.AddChild(newCombineRulesetsListHeader(i18n.Text("Combined Ruleset Files")))

	leftColumn := unison.NewPanel()
	leftColumn.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	leftColumn.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	leftColumn.AddChild(d.filterField)
	leftColumn.AddChild(newCombineRulesetsListScroller(d.left))
	lists.AddChild(leftColumn)

	moveButtons := unison.NewPanel()
	moveButtons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	moveButtons.SetLayoutData(&unison.FlexLayoutData{VAlign: align.Middle})
	d.addButton = unison.NewButton()
	d.addButton.SetTitle(i18n.Text("Add →"))
	d.addButton.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	moveButtons.AddChild(d.addButton)
	d.removeButton = unison.NewButton()
	d.removeButton.SetTitle(i18n.Text("← Remove"))
	d.removeButton.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	moveButtons.AddChild(d.removeButton)
	lists.AddChild(moveButtons)

	rightColumn := unison.NewPanel()
	rightColumn.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	rightColumn.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	rightColumn.AddChild(newCombineRulesetsListScroller(d.right))
	fileButtons := unison.NewPanel()
	fileButtons.SetLayout(&unison.FlexLayout{
		Columns:  5,
		HSpacing: unison.StdHSpacing,
	})
	d.newFileButton = unison.NewButton()
	d.newFileButton.SetTitle(i18n.Text("New File"))
	d.newFileButton.Tooltip = newWrappedTooltip(i18n.Text("Create a new combined file holding the selected components"))
	fileButtons.AddChild(d.newFileButton)
	d.renameButton = unison.NewButton()
	d.renameButton.SetTitle(i18n.Text("Rename"))
	d.renameButton.Tooltip = newWrappedTooltip(i18n.Text("Rename the selected combined file. The extension cannot be changed."))
	fileButtons.AddChild(d.renameButton)
	d.moveButton = unison.NewButton()
	d.moveButton.SetTitle(i18n.Text("Move To…"))
	d.moveButton.Tooltip = newWrappedTooltip(i18n.Text("Move the selected components to another combined file of the same type"))
	fileButtons.AddChild(d.moveButton)
	d.raiseButton = unison.NewButton()
	d.raiseButton.SetTitle(i18n.Text("Raise Priority"))
	fileButtons.AddChild(d.raiseButton)
	d.lowerButton = unison.NewButton()
	d.lowerButton.SetTitle(i18n.Text("Lower Priority"))
	fileButtons.AddChild(d.lowerButton)
	rightColumn.AddChild(fileButtons)
	lists.AddChild(rightColumn)

	panel.AddChild(d.subfoldersCheckBox)

	note := unison.NewLabel()
	note.SetTitle(i18n.Text("Within each combined file, items found in more than one component keep the stats of the highest priority component and gather every component's page references."))
	panel.AddChild(note)

	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon, unison.DefaultDialogTheme.QuestionIconInk,
		panel, []*unison.DialogButtonInfo{
			unison.NewCancelButtonInfo(),
			unison.NewOKButtonInfoWithTitle(i18n.Text("Create")),
		})
	if err != nil {
		errs.Log(err)
		return
	}
	d.okButton = dialog.Button(unison.ModalResponseOK)

	// Drag & drop: both trees can start a drag; only the staging tree accepts drops. Dragging within the staging
	// tree moves; dragging from the source tree copies. The drop rearranges the visible tree, after which the plan
	// is rebuilt from what the tree now shows and re-rendered in canonical form. The dialog's own window must be
	// registered for the drag type, or drops are never delivered to it.
	registerWindowDragTypes(dialog.Window())
	d.left.InstallDragSupport(unison.DocumentSVG, combineDragKey, i18n.Text("Item"), i18n.Text("Items"))
	d.right.InstallDragSupport(unison.DocumentSVG, combineDragKey, i18n.Text("Item"), i18n.Text("Items"))
	d.right.InstallDropSupport(combineDragKey,
		func(from, _ *unison.Table[*combineNode]) bool { return from == d.right },
		nil,
		func(_ *unison.UndoEdit[struct{}], _, _ *unison.Table[*combineNode], _ bool) {
			d.rebuildPlanFromRightTree()
		})

	// The name is validated rather than quietly rewritten, so that what the user typed is what they get. The field
	// paints itself invalid while there is a problem, and its tooltip carries the reason.
	d.nameField.ValidateCallback = func() bool {
		problem := validateCombinedRulesetName(d.nameField.Text())
		if problem == "" {
			d.nameField.Tooltip = nil
		} else {
			d.nameField.Tooltip = newWrappedTooltip(problem)
		}
		return problem == ""
	}
	d.nameField.ModifiedCallback = func(_, _ *unison.FieldState) {
		// Default file names track the name, so the staging must be re-rendered as it changes.
		d.refreshRight(nil)
	}
	d.left.SelectionChangedCallback = d.syncButtons
	d.right.SelectionChangedCallback = d.syncButtons
	d.left.DoubleClickCallback = d.addSelection
	d.addButton.ClickCallback = d.addSelection
	d.removeButton.ClickCallback = d.removeSelection
	d.newFileButton.ClickCallback = d.newFileFromSelection
	d.renameButton.ClickCallback = d.renameSelection
	d.moveButton.ClickCallback = d.moveSelection
	d.raiseButton.ClickCallback = func() { d.shiftComponent(-1) }
	d.lowerButton.ClickCallback = func() { d.shiftComponent(1) }
	d.refreshRight(nil)

	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	name := strings.TrimSpace(d.nameField.Text())
	if name == "" || validateCombinedRulesetName(name) != "" || !d.plan.HasContent() {
		return
	}
	performRulesetCombination(name, d.plan)
}

// syncButtons enables and disables the buttons to match the current selections.
func (d *combineRulesetsDialog) syncButtons() {
	d.addButton.SetEnabled(d.left.SelectionCount() != 0)
	d.removeButton.SetEnabled(d.right.SelectionCount() != 0)
	d.renameButton.SetEnabled(d.right.SelectionCount() == 1 &&
		(d.right.SelectedRows(false)[0].role == combineRoleOutFile ||
			d.right.SelectedRows(false)[0].role == combineRoleOutFolder))
	comps := d.selectedComponents()
	sameExt := len(comps) != 0 && len(comps) == d.right.SelectionCount()
	for _, one := range comps {
		if one.outFile.Ext != comps[0].outFile.Ext {
			sameExt = false
			break
		}
	}
	d.newFileButton.SetEnabled(sameExt)
	d.moveButton.SetEnabled(sameExt && len(d.moveTargetsFor(comps)) != 0)
	single := len(comps) == 1 && d.right.SelectionCount() == 1
	d.raiseButton.SetEnabled(single && comps[0].compIndex > 0)
	d.lowerButton.SetEnabled(single && comps[0].compIndex < len(comps[0].outFile.Components)-1)
	// The Create button needs a valid name and something staged. The field revalidates itself only after the
	// modified callback has run, so force it here to keep the button in step with the current text.
	d.nameField.Validate()
	d.okButton.SetEnabled(!d.nameField.Invalid() && strings.TrimSpace(d.nameField.Text()) != "" && d.plan.HasContent())
}

// selectedComponents returns the right tree's selected component nodes.
func (d *combineRulesetsDialog) selectedComponents() []*combineNode {
	var comps []*combineNode
	for _, row := range d.right.SelectedRows(false) {
		if row.role == combineRoleComponent {
			comps = append(comps, row)
		}
	}
	return comps
}

// moveTargetsFor returns the staged files the selected components could move to: every file of the same type other
// than the one holding the first selected component.
func (d *combineRulesetsDialog) moveTargetsFor(comps []*combineNode) []*gurps.CombinedFile {
	if len(comps) == 0 {
		return nil
	}
	var targets []*gurps.CombinedFile
	for _, f := range d.plan.Files {
		if f.Ext == comps[0].outFile.Ext && f != comps[0].outFile {
			targets = append(targets, f)
		}
	}
	return targets
}

// refreshLeft rebuilds the left tree from the libraries on disk, honoring the filter.
func (d *combineRulesetsDialog) refreshLeft() {
	filter := strings.ToLower(strings.TrimSpace(d.filterField.Text()))
	var roots []*combineNode
	for _, lib := range gurps.GlobalSettings().Libraries.List() {
		libNode := newCombineNode(combineRoleLibrary, "L:"+lib.Key(), lib.Data().Title, d.openMap, nil)
		libNode.lib = lib
		buildCombineSourceFolder(libNode, lib, "", filter)
		if len(libNode.children) != 0 {
			roots = append(roots, libNode)
		}
	}
	setCombineTreeSizer(roots, func() { d.adjustTreeSizeEventually(d.left, &d.adjustLeftPending) })
	d.left.SetRootRows(roots)
	d.left.SizeColumnsToFit(true)
	if d.addButton != nil {
		d.syncButtons()
	}
}

// adjustTreeSizeEventually re-fits a tree's column shortly after a disclosure toggle, once the table has re-synced.
func (d *combineRulesetsDialog) adjustTreeSizeEventually(table *unison.Table[*combineNode], pending *bool) {
	if !*pending {
		*pending = true
		unison.InvokeTaskAfter(func() {
			*pending = false
			table.SyncToModel()
			table.SizeColumnsToFit(true)
		}, time.Millisecond)
	}
}

// buildCombineSourceFolder fills parent with the folders and combinable data files beneath dir (a library-relative
// slash path, "" for the root), pruning branches that hold no combinable files and, when a filter is set, branches
// with no match. A folder that matches the filter keeps everything beneath it.
func buildCombineSourceFolder(parent *combineNode, lib *gurps.Library, dir, filter string) {
	dirOnDisk := lib.Path()
	if dir != "" {
		dirOnDisk = filepath.Join(dirOnDisk, filepath.FromSlash(dir))
	}
	entries, err := os.ReadDir(dirOnDisk)
	if err != nil {
		errs.Log(err, "dir", dirOnDisk)
		return
	}
	names := make([]string, 0, len(entries))
	dirSet := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.IsDir() {
			names = append(names, entry.Name())
			dirSet[entry.Name()] = true
		} else if dir != "" && gurps.CombinableExt(path.Ext(entry.Name())) {
			// Data files directly at a library's root have no source folder and are not offered.
			names = append(names, entry.Name())
		}
	}
	xstrings.SortStringsNaturalAscending(names)
	for _, name := range names {
		rel := path.Join(dir, name)
		matches := filter == "" || strings.Contains(strings.ToLower(name), filter)
		if dirSet[name] {
			folder := newCombineNode(combineRoleSourceFolder, "L:"+lib.Key()+":"+rel, name, parent.openMap, parent)
			folder.lib = lib
			folder.relPath = rel
			subFilter := filter
			if matches {
				subFilter = "" // A matching folder keeps everything beneath it.
			}
			buildCombineSourceFolder(folder, lib, rel, subFilter)
			if len(folder.children) == 0 {
				parent.children = parent.children[:len(parent.children)-1]
				folder.parent = nil
			}
		} else if matches {
			file := newCombineNode(combineRoleSourceFile, "L:"+lib.Key()+":"+rel, name, parent.openMap, parent)
			file.lib = lib
			file.relPath = rel
			file.icon = path.Ext(name)
		}
	}
}

// refreshRight rebuilds the right tree from the plan. When selectKeys is nil the current selection is kept, as far as
// its rows still exist; otherwise the nodes with those keys are selected.
func (d *combineRulesetsDialog) refreshRight(selectKeys []string) {
	if selectKeys == nil {
		for _, row := range d.right.SelectedRows(false) {
			selectKeys = append(selectKeys, row.key)
		}
	}
	name := strings.TrimSpace(d.nameField.Text())
	if name == "" {
		name = i18n.Text("Combined")
	}
	folders := make(map[string]*combineNode)
	var roots []*combineNode
	var folderFor func(subpath string) *combineNode
	folderFor = func(subpath string) *combineNode {
		if subpath == "" {
			return nil
		}
		if n, ok := folders[subpath]; ok {
			return n
		}
		var parent *combineNode
		if i := strings.LastIndex(subpath, "/"); i != -1 {
			parent = folderFor(subpath[:i])
		}
		n := newCombineNode(combineRoleOutFolder, "D:"+strings.ToLower(subpath), path.Base(subpath), d.openMap, parent)
		if parent == nil {
			roots = append(roots, n)
		}
		folders[subpath] = n
		return n
	}
	// Walk the files sorted by subfolder so that parent folders appear before deeper content, keeping the tree's
	// ordering stable no matter the order the plan's slices are in.
	files := slices.Clone(d.plan.Files)
	slices.SortStableFunc(files, func(a, b *gurps.CombinedFile) int {
		if c := xstrings.NaturalCmp(a.Subpath, b.Subpath, true); c != 0 {
			return c
		}
		return xstrings.NaturalCmp(a.FileName(name), b.FileName(name), true)
	})
	for _, f := range files {
		fileName := f.FileName(name)
		parent := folderFor(f.Subpath)
		fileNode := newCombineNode(combineRoleOutFile, "F:"+strings.ToLower(f.Subpath+"/"+fileName), fileName,
			d.openMap, parent)
		fileNode.icon = f.Ext
		fileNode.outFile = f
		if parent == nil {
			roots = append(roots, fileNode)
		}
		for i, comp := range f.Components {
			compNode := newCombineNode(combineRoleComponent, componentKey(comp), path.Base(comp.Path), d.openMap,
				fileNode)
			compNode.icon = comp.Ext()
			compNode.outFile = f
			compNode.compIndex = i
			compNode.comp = comp
			compNode.tooltip = comp.Library.Data().Title + ": " + comp.Path
		}
	}
	setCombineTreeSizer(roots, func() { d.adjustTreeSizeEventually(d.right, &d.adjustRightPending) })
	d.right.SetRootRows(roots)
	d.right.SizeColumnsToFit(true)
	if len(selectKeys) != 0 {
		var indexes []int
		for i := 0; ; i++ {
			row := d.right.RowFromIndex(i)
			if row == nil {
				break
			}
			if slices.Contains(selectKeys, row.key) {
				indexes = append(indexes, i)
			}
		}
		if len(indexes) != 0 {
			d.right.SelectByIndex(indexes...)
		}
	}
	if d.addButton != nil {
		d.syncButtons()
	}
}

// addSelection stages the selected left rows at their default locations: a library stages all of its folders, a
// top-level folder stages its files (and, when the checkbox is on, its subfolders'), a deeper folder stages
// everything beneath it, and a file stages just itself. Files land in the output at the same relative position they
// hold within their source folder.
func (d *combineRulesetsDialog) addSelection() {
	includeSubfolders := d.subfoldersCheckBox.State == check.On
	for _, row := range d.left.SelectedRows(true) {
		switch row.role {
		case combineRoleLibrary:
			for _, child := range row.children {
				if child.role == combineRoleSourceFolder {
					d.addFolder(child, includeSubfolders)
				}
			}
		case combineRoleSourceFolder:
			d.addFolder(row, includeSubfolders)
		case combineRoleSourceFile:
			folder, rest := topFolderOf(row.relPath)
			d.plan.AddFile(gurps.CombinedLibrarySource{Library: row.lib, Folder: folder}, rest)
		default:
		}
	}
	d.refreshRight(nil)
}

// addFolder stages one folder node at its default locations.
func (d *combineRulesetsDialog) addFolder(node *combineNode, includeSubfolders bool) {
	stageCombineFolder(d.plan, node.lib, node.relPath, includeSubfolders)
}

// gatherCombineFolderFiles lists every file beneath dirOnDisk, relative to it in slash form, skipping hidden entries.
func gatherCombineFolderFiles(dirOnDisk string) ([]string, error) {
	var rel []string
	if err := filepath.WalkDir(dirOnDisk, func(p string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() && p != dirOnDisk {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.IsDir() {
			sub, relErr := filepath.Rel(dirOnDisk, p)
			if relErr != nil {
				return relErr
			}
			rel = append(rel, filepath.ToSlash(sub))
		}
		return nil
	}); err != nil {
		return nil, errs.Wrap(err)
	}
	xstrings.SortStringsNaturalAscending(rel)
	return rel, nil
}

// removeSelection removes the selected right rows from the staging: components leave their files, files leave the
// plan, and folders take everything beneath them.
func (d *combineRulesetsDialog) removeSelection() {
	type removal struct {
		file *gurps.CombinedFile
		comp gurps.CombinedComponent
	}
	var compRemovals []removal
	for _, row := range d.right.SelectedRows(true) {
		switch row.role {
		case combineRoleComponent:
			compRemovals = append(compRemovals, removal{file: row.outFile, comp: row.outFile.Components[row.compIndex]})
		case combineRoleOutFile:
			d.plan.Files = slices.DeleteFunc(d.plan.Files, func(f *gurps.CombinedFile) bool { return f == row.outFile })
		case combineRoleOutFolder:
			subpath := row.key[len("D:"):]
			d.plan.Files = slices.DeleteFunc(d.plan.Files, func(f *gurps.CombinedFile) bool {
				lower := strings.ToLower(f.Subpath)
				return lower == subpath || strings.HasPrefix(lower, subpath+"/")
			})
		default:
		}
	}
	for _, one := range compRemovals {
		one.file.Components = slices.DeleteFunc(one.file.Components, func(c gurps.CombinedComponent) bool {
			return c.Library.Key() == one.comp.Library.Key() && strings.EqualFold(c.Path, one.comp.Path)
		})
	}
	d.plan.DropEmptyDefaultFiles()
	d.refreshRight([]string{})
}

// promptForCombinedFileName asks for a file name, showing the immutable extension beside the entry. It returns the
// trimmed name and true when the user confirms a valid one.
func promptForCombinedFileName(title, current, ext string) (string, bool) {
	name := current
	field := NewStringField(nil, "", title, func() string { return name }, func(s string) { name = s })
	field.SetMinimumTextWidthUsing(minTextWidthCandidate)
	field.ValidateCallback = func() bool {
		trimmed := strings.TrimSpace(name)
		return trimmed != "" && validateCombinedRulesetName(trimmed) == ""
	}
	columns := 2
	if ext != "" {
		columns = 3
	}
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  columns,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.AddChild(NewFieldLeadingLabel(title, false))
	panel.AddChild(field)
	if ext != "" {
		extLabel := unison.NewLabel()
		extLabel.SetTitle(ext)
		panel.AddChild(extLabel)
	}
	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon, unison.DefaultDialogTheme.QuestionIconInk,
		panel, []*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		errs.Log(err)
		return "", false
	}
	field.RequestFocus()
	field.SelectAll()
	if dialog.RunModal() != unison.ModalResponseOK {
		return "", false
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || validateCombinedRulesetName(trimmed) != "" {
		return "", false
	}
	return trimmed, true
}

// newFileFromSelection creates a new combined file holding the selected components. The file lands in the same
// subfolder as the file the first component came from.
func (d *combineRulesetsDialog) newFileFromSelection() {
	comps := d.selectedComponents()
	if len(comps) == 0 {
		return
	}
	name, ok := promptForCombinedFileName(i18n.Text("New File Name"), "", comps[0].outFile.Ext)
	if !ok {
		return
	}
	target := &gurps.CombinedFile{
		Subpath:    comps[0].outFile.Subpath,
		CustomName: name,
		Ext:        comps[0].outFile.Ext,
	}
	d.plan.Files = append(d.plan.Files, target)
	d.moveComponents(comps, target)
}

// renameSelection renames the selected combined file -- pinning its name so that it no longer tracks the
// combination's name -- or the selected output folder, rehoming everything beneath it.
func (d *combineRulesetsDialog) renameSelection() {
	rows := d.right.SelectedRows(false)
	if len(rows) != 1 {
		return
	}
	switch rows[0].role {
	case combineRoleOutFile:
		f := rows[0].outFile
		current := strings.TrimSuffix(f.FileName(strings.TrimSpace(d.nameField.Text())), f.Ext)
		name, ok := promptForCombinedFileName(i18n.Text("File Name"), current, f.Ext)
		if !ok {
			return
		}
		f.CustomName = name
	case combineRoleOutFolder:
		oldSub := outFolderSubpath(rows[0])
		name, ok := promptForCombinedFileName(i18n.Text("Folder Name"), path.Base(oldSub), "")
		if !ok {
			return
		}
		newSub := name
		if parent := path.Dir(oldSub); parent != "." {
			newSub = parent + "/" + name
		}
		renameCombineSubpath(d.plan, oldSub, newSub)
	default:
		return
	}
	d.refreshRight(nil)
}

// outFolderSubpath reconstructs an output folder node's subpath from its ancestry.
func outFolderSubpath(n *combineNode) string {
	if n.parent == nil {
		return n.title
	}
	return outFolderSubpath(n.parent) + "/" + n.title
}

// renameCombineSubpath rehomes every staged file under oldSub to the same position under newSub.
func renameCombineSubpath(plan *gurps.CombinedPlan, oldSub, newSub string) {
	for _, f := range plan.Files {
		switch {
		case strings.EqualFold(f.Subpath, oldSub):
			f.Subpath = newSub
		case len(f.Subpath) > len(oldSub) && strings.EqualFold(f.Subpath[:len(oldSub)], oldSub) &&
			f.Subpath[len(oldSub)] == '/':
			f.Subpath = newSub + f.Subpath[len(oldSub):]
		}
	}
}

// moveSelection moves the selected components to another combined file of the same type, chosen from a popup.
func (d *combineRulesetsDialog) moveSelection() {
	comps := d.selectedComponents()
	targets := d.moveTargetsFor(comps)
	if len(targets) == 0 {
		return
	}
	name := strings.TrimSpace(d.nameField.Text())
	popup := unison.NewPopupMenu[string]()
	for _, f := range targets {
		title := f.FileName(name)
		if f.Subpath != "" {
			title = f.Subpath + "/" + title
		}
		popup.AddItem(title)
	}
	popup.SelectIndex(0)
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.AddChild(NewFieldLeadingLabel(i18n.Text("Move to"), false))
	panel.AddChild(popup)
	dialog, err := unison.NewDialog(unison.DefaultDialogTheme.QuestionIcon, unison.DefaultDialogTheme.QuestionIconInk,
		panel, []*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()})
	if err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() != unison.ModalResponseOK {
		return
	}
	index := popup.SelectedIndex()
	if index < 0 || index >= len(targets) {
		return
	}
	d.moveComponents(comps, targets[index])
}

// moveComponents removes the given component nodes from their files and appends them, in the order shown, to target.
func (d *combineRulesetsDialog) moveComponents(comps []*combineNode, target *gurps.CombinedFile) {
	keys := make([]string, 0, len(comps))
	for _, one := range comps {
		comp := one.outFile.Components[one.compIndex]
		one.outFile.Components = slices.DeleteFunc(one.outFile.Components, func(c gurps.CombinedComponent) bool {
			return c.Library.Key() == comp.Library.Key() && strings.EqualFold(c.Path, comp.Path)
		})
		target.Components = append(target.Components, comp)
		keys = append(keys, componentKey(comp))
	}
	d.plan.DropEmptyDefaultFiles()
	d.refreshRight(keys)
}

// rebuildPlanFromRightTree turns whatever arrangement a drop produced back into the plan, then re-renders the tree in
// canonical form.
func (d *combineRulesetsDialog) rebuildPlanFromRightTree() {
	rebuildCombinePlan(d.plan, d.right.RootRows(), d.subfoldersCheckBox.State == check.On)
	d.refreshRight(nil)
}

// rebuildCombinePlan reads the staging back out of the tree: each file's components are the component nodes beneath
// it, in the order shown; a file sitting under a different folder than before has been rehomed there. Nodes dragged
// in from the source tree become staged content: a source file dropped inside a matching file joins it right there,
// while anything else lands at its default location. A component that ends up somewhere it cannot live -- outside any
// file, or in a file of another type -- snaps back to the file it came from.
func rebuildCombinePlan(plan *gurps.CombinedPlan, roots []*combineNode, includeSubfolders bool) {
	type snapBack struct {
		file *gurps.CombinedFile
		comp gurps.CombinedComponent
	}
	newComps := make(map[*gurps.CombinedFile][]gurps.CombinedComponent)
	seen := make(map[string]bool)
	var orphans []snapBack
	var deferredNodes []*combineNode
	place := func(file *gurps.CombinedFile, comp gurps.CombinedComponent) {
		key := componentKey(comp)
		if seen[key] {
			return
		}
		seen[key] = true
		newComps[file] = append(newComps[file], comp)
	}
	var walk func(nodes []*combineNode, curFile *gurps.CombinedFile, curSubpath string)
	walk = func(nodes []*combineNode, curFile *gurps.CombinedFile, curSubpath string) {
		for _, n := range nodes {
			switch n.role {
			case combineRoleOutFolder:
				sub := n.title
				if curSubpath != "" {
					sub = curSubpath + "/" + n.title
				}
				walk(n.children, nil, sub)
			case combineRoleOutFile:
				n.outFile.Subpath = curSubpath
				walk(n.children, n.outFile, curSubpath)
			case combineRoleComponent:
				if curFile != nil && curFile.Ext == n.comp.Ext() {
					place(curFile, n.comp)
				} else if n.outFile != nil {
					orphans = append(orphans, snapBack{file: n.outFile, comp: n.comp})
				}
			case combineRoleSourceFile:
				comp := gurps.CombinedComponent{Library: n.lib, Path: n.relPath}
				if curFile != nil && curFile.Ext == comp.Ext() {
					if !seen[componentKey(comp)] && !plan.Contains(comp) {
						place(curFile, comp)
					}
				} else {
					deferredNodes = append(deferredNodes, n)
				}
			case combineRoleSourceFolder, combineRoleLibrary:
				deferredNodes = append(deferredNodes, n)
			default:
			}
		}
	}
	walk(roots, nil, "")
	for _, f := range plan.Files {
		f.Components = newComps[f]
	}
	for _, one := range orphans {
		if !seen[componentKey(one.comp)] {
			seen[componentKey(one.comp)] = true
			one.file.Components = append(one.file.Components, one.comp)
		}
	}
	// Source folders, libraries, and files that were not dropped into a matching file stage at their defaults.
	for _, n := range deferredNodes {
		switch n.role {
		case combineRoleSourceFile:
			folder, rest := topFolderOf(n.relPath)
			plan.AddFile(gurps.CombinedLibrarySource{Library: n.lib, Folder: folder}, rest)
		case combineRoleSourceFolder:
			stageCombineFolder(plan, n.lib, n.relPath, includeSubfolders)
		case combineRoleLibrary:
			for _, child := range n.children {
				if child.role == combineRoleSourceFolder {
					stageCombineFolder(plan, child.lib, child.relPath, includeSubfolders)
				}
			}
		default:
		}
	}
	plan.DropEmptyDefaultFiles()
}

// stageCombineFolder stages one source folder at its default locations: a top-level folder as a source in its own
// right, a deeper folder as every combinable file beneath it, at the position each holds within its top-level folder.
func stageCombineFolder(plan *gurps.CombinedPlan, lib *gurps.Library, relPath string, includeSubfolders bool) {
	folder, rest := topFolderOf(relPath)
	src := gurps.CombinedLibrarySource{Library: lib, Folder: folder}
	if rest == "" {
		if err := plan.AddSource(src, includeSubfolders); err != nil {
			Workspace.ErrorHandler(i18n.Text("Unable to add the folder"), err)
		}
		return
	}
	dirOnDisk := filepath.Join(lib.Path(), filepath.FromSlash(relPath))
	rel, err := gatherCombineFolderFiles(dirOnDisk)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to add the folder"), err)
		return
	}
	for _, one := range rel {
		plan.AddFile(src, path.Join(rest, one))
	}
}

// shiftComponent moves the selected component up or down within its file, changing its priority.
func (d *combineRulesetsDialog) shiftComponent(delta int) {
	comps := d.selectedComponents()
	if len(comps) != 1 {
		return
	}
	f := comps[0].outFile
	i := comps[0].compIndex
	j := i + delta
	if j < 0 || j >= len(f.Components) {
		return
	}
	f.Components[i], f.Components[j] = f.Components[j], f.Components[i]
	d.refreshRight([]string{componentKey(f.Components[j])})
}

// combinedRulesetInvalidNameChars are the characters Windows refuses in a file or folder name. The forward slash is
// also a separator everywhere else.
const combinedRulesetInvalidNameChars = `<>:"/\|?*`

// combinedRulesetReservedNames are the device names Windows will not allow a file or folder to be called, with or
// without an extension.
var combinedRulesetReservedNames = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true, "com5": true,
	"com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true,
	"lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// validateCombinedRulesetName returns a message explaining why name cannot be used as a file or folder name, or an
// empty string when it is usable. Windows' rules -- the strictest of the platforms GCS runs on -- are applied
// everywhere, so that a combined library created on one platform stays usable on the others. An empty name reports no
// problem, since the buttons gated on the name are already disabled for it and nagging about a field that has not
// been filled in yet is noise.
func validateCombinedRulesetName(name string) string {
	trimmed := strings.TrimSpace(name)
	switch {
	case trimmed == "":
		return ""
	case strings.ContainsAny(trimmed, combinedRulesetInvalidNameChars):
		return fmt.Sprintf(i18n.Text("A name may not contain any of these characters: %s"),
			combinedRulesetInvalidNameChars)
	case strings.ContainsFunc(trimmed, func(r rune) bool { return r < ' ' || r == 0x7f }):
		return i18n.Text("A name may not contain control characters")
	case trimmed == "." || trimmed == "..":
		return i18n.Text("A name may not be \".\" or \"..\"")
	case strings.HasSuffix(trimmed, "."):
		return i18n.Text("A name may not end with a period")
	case combinedRulesetReservedNames[strings.ToLower(strings.TrimSuffix(trimmed, filepath.Ext(trimmed)))]:
		return fmt.Sprintf(i18n.Text("%s is a reserved name"), trimmed)
	default:
		return ""
	}
}

// performRulesetCombination runs the combination for the staged plan, writing the result into a folder in the user
// library named for the combination.
func performRulesetCombination(name string, plan *gurps.CombinedPlan) {
	unableMsg := i18n.Text("Unable to create the combined ruleset")
	destDir := filepath.Join(gurps.GlobalSettings().Libraries.User().Path(), name)
	if xos.IsDir(destDir) {
		if unison.YesNoDialog(fmt.Sprintf(i18n.Text(`A folder named "%s" already exists in the User Library.`), name),
			i18n.Text("Any combined data files it contains with the same names will be replaced. Continue?")) != unison.ModalResponseOK {
			return
		}
	}
	// A combined file that is already open would be overwritten out from under its view, so close such views first,
	// the same way the rules lookup download does.
	for _, target := range plan.Targets(name, destDir) {
		target = filepath.Clean(target)
		for _, one := range AllDockables() {
			if tc, ok := one.(unison.TabCloser); ok {
				var fbd FileBackedDockable
				if fbd, ok = one.(FileBackedDockable); ok {
					if filepath.Clean(fbd.BackingFilePath()) == target {
						if !tc.MayAttemptClose() || !tc.AttemptClose() {
							unison.WarningDialogWithMessage(i18n.Text("Combination canceled"),
								i18n.Text("Cannot replace a combined file while it is open."))
							return
						}
						break
					}
				}
			}
		}
	}
	if err := runRulesetCombination(name, plan, destDir); err != nil {
		Workspace.ErrorHandler(unableMsg, err)
		return
	}
	Workspace.Navigator.EventuallyReload()
}

// runRulesetCombination performs the combination off the UI thread, putting up a modal progress window while it runs,
// so that a large set of staged files does not freeze the application.
func runRulesetCombination(name string, plan *gurps.CombinedPlan, destDir string) error {
	frame := windowPlacementFrame()
	wnd, err := unison.NewWindow(i18n.Text("Combining…"), unison.FloatingWindowOption(),
		unison.NotResizableWindowOption(), unison.UndecoratedWindowOption(), unison.TransientWindowOption())
	if err != nil {
		// Better to do the work without the progress window than to refuse to do it at all.
		errs.Log(err)
		_, err = plan.Create(name, destDir)
		return err
	}
	content := unison.NewPanel()
	content.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.NewUniformInsets(1), false), unison.NewEmptyBorder(geom.NewUniformInsets(2*unison.StdHSpacing))))
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: unison.StdVSpacing,
	})
	label := unison.NewLabel()
	label.SetTitle(i18n.Text("Combining the staged files…"))
	content.AddChild(label)
	progress := unison.NewProgressBar(0)
	progress.SetLayoutData(&unison.FlexLayoutData{
		MinSize: geom.Size{Width: 500},
		HAlign:  align.Fill,
		HGrab:   true,
	})
	content.AddChild(progress)
	wnd.SetContent(content)
	wnd.Pack()
	wndFrame := wnd.FrameRect()
	frame.Y += (frame.Height - wndFrame.Height) / 3
	frame.Height = wndFrame.Height
	frame.X += (frame.Width - wndFrame.Width) / 2
	frame.Width = wndFrame.Width
	frame = frame.Align()
	wnd.SetFrameRect(unison.BestDisplayForRect(frame).FitRectOnto(frame))
	wnd.ToFront()
	resultChan := make(chan error, 1)
	go func() {
		_, createErr := plan.Create(name, destDir)
		resultChan <- createErr
		unison.InvokeTask(func() { wnd.StopModal(unison.ModalResponseOK) })
	}()
	wnd.RunModal()
	return <-resultChan
}

func newCombineRulesetsListHeader(title string) *unison.Label {
	header := unison.NewLabel()
	header.SetTitle(title)
	return header
}

func newCombineRulesetsListScroller(content unison.Paneler) *unison.ScrollPanel {
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	scroll.SetContent(content, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: combineTreeSize,
		MinSize:  combineTreeSize,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		HGrab:    true,
		VGrab:    true,
	})
	return scroll
}
