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
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/autoscale"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/check"
)

// PathToLog is set by the main entry point to whatever is being used for the path to the log file.
var PathToLog string

var languageSetting string

type generalSettingsDockable struct {
	SettingsDockable
	nameField                    *StringField
	checkBoxes                   []settingCheckBox
	appUpdateCheckPopup          *unison.PopupMenu[updatecheck.Option]
	libraryUpdateCheckPopup      *unison.PopupMenu[updatecheck.Option]
	deepSearchableCheckbox       []membershipCheckBox[string]
	openInWindowCheckbox         []membershipCheckBox[dgroup.Group]
	pointsField                  *DecimalField
	techLevelField               *StringField
	calendarPopup                *unison.PopupMenu[string]
	initialScaleFields           []initialScaleField
	autoScalingPopup             *unison.PopupMenu[autoscale.Option]
	maxAutoColWidthField         *IntegerField
	monitorResolutionField       *IntegerField
	exportResolutionField        *IntegerField
	permittedScriptExecTimeField *DecimalField
	tooltipDelayField            *DecimalField
	tooltipDismissalField        *DecimalField
	scrollWheelMultiplierField   *DecimalField
	cursorSizeField              *IntegerField
	externalPDFCmdlineField      *StringField
	localeField                  *StringField
}

// settingCheckBox pairs a checkbox with the general setting it edits, so that sync() can bring it back into line.
type settingCheckBox struct {
	box   *CheckBox
	value *bool
}

// initialScaleField pairs an "Initial ... Scale" field with the general setting it edits, so that sync() can bring it
// back into line.
type initialScaleField struct {
	field *PercentageField
	value *int
}

// membershipCheckBox pairs a checkbox with the item whose membership in a settings list it toggles.
type membershipCheckBox[T cmp.Ordered] struct {
	box  *CheckBox
	item T
}

// ShowGeneralSettings shows the General Settings window.
func ShowGeneralSettings() {
	if activateDockable[*generalSettingsDockable](nil) {
		return
	}
	d := &generalSettingsDockable{}
	d.initSettings(d, &settingsSpec{
		title:             i18n.Text("General Settings"),
		ext:               gurps.GeneralSettingsExt,
		loader:            d.load,
		saver:             d.save,
		resetter:          d.reset,
		willClose:         d.willClose,
		addToStartToolbar: d.addToStartToolbar,
		initContent:       d.initContent,
	})
	d.nameField.RequestFocus()
}

func (d *generalSettingsDockable) addToStartToolbar(toolbar *unison.Panel) {
	addHelpButton(toolbar, "md:User%20Guide/General%20Settings")
}

func (d *generalSettingsDockable) initContent(content *unison.Panel) {
	initSettingsContent(content, 3)
	d.createPlayerAndDescFields(content)
	d.createCheckboxBlock(content)
	d.createUpdateCheckPopups(content)
	d.createInitialPointsFields(content)
	d.createTechLevelField(content)
	d.createCalendarPopup(content)
	d.createInitialScaleFields(content)
	d.createCellAutoMaxWidthField(content)
	d.createMonitorResolutionField(content)
	d.createImageResolutionField(content)
	d.createCursorSizeField(content)
	d.createPermittedScriptExecTimeField(content)
	d.createTooltipDelayField(content)
	d.createTooltipDismissalField(content)
	d.createScrollWheelMultiplierField(content)
	d.createPathInfoField(content, i18n.Text("Settings Path"), gurps.SettingsPath)
	d.createPathInfoField(content, i18n.Text("Translations Path"), i18n.Dir)
	d.createPathInfoField(content, i18n.Text("Log Path"), PathToLog)
	d.createExternalPDFCmdLineField(content)
	d.createLocaleField(content)
	d.createDeepSearchCheckboxes(content)
	d.createOpenInWindowCheckboxes(content)
}

func (d *generalSettingsDockable) createPlayerAndDescFields(content *unison.Panel) {
	title := i18n.Text("Default Player Name")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.nameField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().General.DefaultPlayerName },
		func(s string) { gurps.GlobalSettings().General.DefaultPlayerName = s })
	d.nameField.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	content.AddChild(d.nameField)
}

func (d *generalSettingsDockable) createCheckboxBlock(content *unison.Panel) {
	gs := gurps.GlobalSettings().General
	d.addGeneralCheckBox(content, i18n.Text("Restore workspace arrangement on start"), &gs.RestoreWorkspaceOnStart, nil)
	d.addGeneralCheckBox(content, i18n.Text("Fill in initial description"), &gs.AutoFillProfile, nil)
	d.addGeneralCheckBox(content, i18n.Text("Group containers when sorting"), &gs.GroupContainersOnSort,
		func() { Workspace.Navigator.EventuallyReload() })
	d.addGeneralCheckBox(content, i18n.Text("Add natural attacks to new sheets"), &gs.AutoAddNaturalAttacks, nil)
	d.addGeneralCheckBox(content, i18n.Text("Initial click on text field selects all"),
		&gs.InitialFieldClickSelectsAll, nil)
	box := d.addGeneralCheckBox(content, i18n.Text("Show additional page references when space allows"),
		&gs.ExpandPageReferences, func() {
			for _, one := range AllDockables() {
				if r, ok := one.(Rebuildable); ok {
					r.Rebuild(false)
				}
			}
		})
	box.Tooltip = newWrappedTooltip(i18n.Text(`When enabled, the page reference column on character sheets, loot sheets, and templates will display more than one page reference when there is room, rather than always collapsing to a single reference. Standalone lists are unaffected, since their columns can be resized directly.`))
}

// addGeneralCheckBox adds a checkbox for the general setting that value points at, keeping it in the field column of
// the content by preceding it with an empty label. onChange, if not nil, runs after the setting has been changed
// through the checkbox. The pointer is stable because reset and load copy into the settings rather than replacing them.
func (d *generalSettingsDockable) addGeneralCheckBox(content *unison.Panel, title string, value *bool, onChange func()) *CheckBox {
	content.AddChild(NewFieldLeadingLabel("", false))
	box := addCheckBox(content, title, value)
	box.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	box.OnSet = onChange
	d.checkBoxes = append(d.checkBoxes, settingCheckBox{box: box, value: value})
	return box
}

func (d *generalSettingsDockable) createUpdateCheckPopups(content *unison.Panel) {
	d.appUpdateCheckPopup = newUpdateCheckPopup(content,
		fmt.Sprintf(i18n.Text("Check for %s Updates"), xos.AppName),
		fmt.Sprintf(i18n.Text("How often to look for a newer version of %s. Help ▸ Check for %s updates always works."),
			xos.AppName, xos.AppName),
		func() updatecheck.Option { return gurps.GlobalSettings().General.AppUpdateCheck },
		func(option updatecheck.Option) { gurps.GlobalSettings().General.AppUpdateCheck = option })
	d.libraryUpdateCheckPopup = newUpdateCheckPopup(content, i18n.Text("Check for Library Updates"),
		i18n.Text("How often to look for newer versions of the libraries. The Library Explorer's update buttons always work, checking for releases when clicked if no check has been made yet."),
		func() updatecheck.Option { return gurps.GlobalSettings().General.LibraryUpdateCheck },
		func(option updatecheck.Option) { gurps.GlobalSettings().General.LibraryUpdateCheck = option })
}

// newUpdateCheckPopup adds a labeled popup menu for choosing how often an update check should be made.
func newUpdateCheckPopup(content *unison.Panel, title, tooltip string, get func() updatecheck.Option,
	set func(updatecheck.Option),
) *unison.PopupMenu[updatecheck.Option] {
	content.AddChild(NewFieldLeadingLabel(title, false))
	popup := newPopupMenu(updatecheck.Options, get(), func(option updatecheck.Option) {
		set(option)
		ApplyUpdateCheckSettings()
	})
	popup.Tooltip = newWrappedTooltip(tooltip)
	popup.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	content.AddChild(popup)
	return popup
}

func (d *generalSettingsDockable) createInitialPointsFields(content *unison.Panel) {
	title := i18n.Text("Initial Points")
	d.pointsField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.InitialPoints },
		func(v fxp.Int) { gurps.GlobalSettings().General.InitialPoints = v }, gurps.InitialPointsMin,
		gurps.InitialPointsMax, false, false)
	addLabeledSettingField(content, title, d.pointsField)
}

func (d *generalSettingsDockable) createTechLevelField(content *unison.Panel) {
	title := i18n.Text("Default Tech Level")
	d.techLevelField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().General.DefaultTechLevel },
		func(s string) { gurps.GlobalSettings().General.DefaultTechLevel = s })
	d.techLevelField.Tooltip = newWrappedTooltip(gurps.TechLevelInfo())
	d.techLevelField.SetMinimumTextWidthUsing("12^")
	addLabeledSettingField(content, title, d.techLevelField)
}

func (d *generalSettingsDockable) createCalendarPopup(content *unison.Panel) {
	d.calendarPopup = unison.NewPopupMenu[string]()
	libraries := gurps.GlobalSettings().Libraries
	for _, lib := range gurps.AvailableCalendarRefs(libraries) {
		d.calendarPopup.AddDisabledItem(lib.Name)
		for _, one := range lib.List {
			d.calendarPopup.AddItem(one.Name)
		}
	}
	installPopupSelection(d.calendarPopup, gurps.GlobalSettings().General.CalendarRef(libraries).Name,
		func(name string) { gurps.GlobalSettings().General.CalendarName = name })
	addLabeledSettingField(content, i18n.Text("Calendar"), d.calendarPopup)
}

func (d *generalSettingsDockable) createInitialScaleFields(content *unison.Panel) {
	gs := gurps.GlobalSettings().General
	d.addInitialScaleField(content, i18n.Text("Initial List Scale"), &gs.InitialListUIScale)
	d.addInitialScaleField(content, i18n.Text("Initial Editor Scale"), &gs.InitialEditorUIScale)
	d.addInitialScaleField(content, i18n.Text("Initial Sheet Scale"), &gs.InitialSheetUIScale)
	d.autoScalingPopup = newPopupMenu(autoscale.Options, gs.PDFAutoScaling,
		func(mode autoscale.Option) { gurps.GlobalSettings().General.PDFAutoScaling = mode })
	d.addInitialScaleField(content, i18n.Text("Initial PDF Scale"), &gs.InitialPDFUIScale, d.autoScalingPopup)
	d.addInitialScaleField(content, i18n.Text("Initial Markdown Scale"), &gs.InitialMarkdownUIScale)
	d.addInitialScaleField(content, i18n.Text("Initial Image Scale"), &gs.InitialImageUIScale)
}

// addInitialScaleField adds a labeled percentage field for the initial UI scale setting that value points at, followed
// on the same row by any trailing panels. The pointer is stable because reset and load copy into the settings rather
// than replacing them.
func (d *generalSettingsDockable) addInitialScaleField(content *unison.Panel, title string, value *int, trailing ...unison.Paneler) *PercentageField {
	field := NewPercentageField(nil, "", title, func() int { return *value }, func(v int) { *value = v },
		gurps.InitialUIScaleMin, gurps.InitialUIScaleMax, false, false)
	addLabeledSettingField(content, title, field, trailing...)
	d.initialScaleFields = append(d.initialScaleFields, initialScaleField{field: field, value: value})
	return field
}

func (d *generalSettingsDockable) createCellAutoMaxWidthField(content *unison.Panel) {
	title := i18n.Text("Max Auto Column Width")
	d.maxAutoColWidthField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.MaximumAutoColWidth },
		func(v int) { gurps.GlobalSettings().General.MaximumAutoColWidth = v },
		gurps.AutoColWidthMin, gurps.AutoColWidthMax, false, false)
	addLabeledSettingField(content, title, d.maxAutoColWidthField)
}

func (d *generalSettingsDockable) createMonitorResolutionField(content *unison.Panel) {
	title := i18n.Text("Monitor Resolution")
	d.monitorResolutionField = NewNumericFieldWithException(nil, "", title,
		func(minValue, maxValue int) []int { return []int{minValue, maxValue} },
		func() int { return gurps.GlobalSettings().General.MonitorResolution },
		func(v int) { gurps.GlobalSettings().General.MonitorResolution = v },
		strconv.Itoa, strconv.Atoi, gurps.MonitorResolutionMin, gurps.MonitorResolutionMax, 0)
	addLabeledSettingField(content, title, d.monitorResolutionField,
		NewFieldTrailingLabel(i18n.Text("ppi (A value of 0 will cause the ppi reported by your monitor to be used)"), false))
}

func (d *generalSettingsDockable) createImageResolutionField(content *unison.Panel) {
	title := i18n.Text("Image Export Resolution")
	d.exportResolutionField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.ImageResolution },
		func(v int) { gurps.GlobalSettings().General.ImageResolution = v },
		gurps.ImageResolutionMin, gurps.ImageResolutionMax, false, false)
	addLabeledSettingField(content, title, d.exportResolutionField, NewFieldTrailingLabel(i18n.Text("ppi"), false))
}

func (d *generalSettingsDockable) createPermittedScriptExecTimeField(content *unison.Panel) {
	title := i18n.Text("Max Execution Time")
	d.permittedScriptExecTimeField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.PermittedPerScriptExecTime },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.PermittedPerScriptExecTime = v
			gurps.SyncScriptExecTimeLimit()
		}, gurps.PermittedScriptExecTimeMin, gurps.PermittedScriptExecTimeMax, false, false)
	addLabeledSettingField(content, title, d.permittedScriptExecTimeField,
		NewFieldTrailingLabel(i18n.Text("seconds per script"), false))
}

func (d *generalSettingsDockable) createTooltipDelayField(content *unison.Panel) {
	title := i18n.Text("Tooltip Delay")
	d.tooltipDelayField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.TooltipDelay },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.TooltipDelay = v
			general.UpdateToolTipTiming()
		}, gurps.TooltipDelayMin, gurps.TooltipDelayMax, false, false)
	addLabeledSettingField(content, title, d.tooltipDelayField, NewFieldTrailingLabel(i18n.Text("seconds"), false))
}

func (d *generalSettingsDockable) createTooltipDismissalField(content *unison.Panel) {
	title := i18n.Text("Tooltip Dismissal")
	d.tooltipDismissalField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.TooltipDismissal },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.TooltipDismissal = v
			general.UpdateToolTipTiming()
		}, gurps.TooltipDismissalMin, gurps.TooltipDismissalMax, false, false)
	addLabeledSettingField(content, title, d.tooltipDismissalField, NewFieldTrailingLabel(i18n.Text("seconds"), false))
}

func (d *generalSettingsDockable) createScrollWheelMultiplierField(content *unison.Panel) {
	title := i18n.Text("Scroll Wheel Multiplier")
	d.scrollWheelMultiplierField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.ScrollWheelMultiplier },
		func(v fxp.Int) { gurps.GlobalSettings().General.ScrollWheelMultiplier = v },
		gurps.ScrollWheelMultiplierMin, gurps.ScrollWheelMultiplierMax, false, false)
	addLabeledSettingField(content, title, d.scrollWheelMultiplierField)
}

func (d *generalSettingsDockable) createCursorSizeField(content *unison.Panel) {
	title := i18n.Text("Cursor Size")
	d.cursorSizeField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.CursorSize },
		func(v int) {
			general := gurps.GlobalSettings().General
			general.CursorSize = v
			general.UpdateCursorSize()
		}, gurps.CursorSizeMin, gurps.CursorSizeMax, false, false)
	addLabeledSettingField(content, title, d.cursorSizeField, NewFieldTrailingLabel(i18n.Text("points"), false))
}

// addLabeledSettingField adds a row to the three-column content: a leading label with the title, then the field
// spanning the remaining two columns. Any trailing panels share the field's span, following it on the same row.
func addLabeledSettingField(content *unison.Panel, title string, field unison.Paneler, trailing ...unison.Paneler) {
	content.AddChild(NewFieldLeadingLabel(title, false))
	if len(trailing) == 0 {
		field.AsPanel().SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
		content.AddChild(field)
		return
	}
	content.AddChild(WrapWithSpan(2, append([]unison.Paneler{field}, trailing...)...))
}

func (d *generalSettingsDockable) createPathInfoField(content *unison.Panel, title, value string) {
	content.AddChild(NewFieldLeadingLabel(title, false))
	content.AddChild(NewNonEditableField(func(field *NonEditableField) {
		field.SetTitle(value)
	}))
	addButton := unison.NewSVGButton(svg.Copy)
	addButton.Tooltip = newWrappedTooltip(i18n.Text("Copy to clipboard"))
	addButton.ClickCallback = func() {
		unison.ClipboardSetText(value)
	}
	content.AddChild(addButton)
}

func (d *generalSettingsDockable) createExternalPDFCmdLineField(content *unison.Panel) {
	title := i18n.Text("External PDF Viewer")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.externalPDFCmdlineField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().General.ExternalPDFCmdLine },
		func(s string) { gurps.GlobalSettings().General.ExternalPDFCmdLine = strings.TrimSpace(s) })
	d.externalPDFCmdlineField.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	d.externalPDFCmdlineField.ValidateCallback = func() bool {
		_, err := xflag.SplitCommandLineWithoutEscapes(strings.TrimSpace(d.externalPDFCmdlineField.Text()))
		return err == nil
	}
	d.externalPDFCmdlineField.Tooltip = newWrappedTooltip(i18n.Text(`The internal PDF viewer will be used if the External PDF Viewer field is empty.
Use $FILE where the full path to the PDF should be placed.
Use $TEXT where the highlighted string should be placed.
Use $PAGE where the page number should be placed.

In most cases, you'll want to surround $FILE and $TEXT with quotes.
Note that this might still fail, e.g. if the variable itself contains quotes.`))
	content.AddChild(d.externalPDFCmdlineField)
}

func (d *generalSettingsDockable) createLocaleField(content *unison.Panel) {
	title := i18n.Text("Interface Locale")
	content.AddChild(NewFieldLeadingLabel(title, false))
	d.localeField = NewStringField(nil, "", title,
		func() string { return languageSetting },
		func(s string) { languageSetting = strings.TrimSpace(s) })
	d.localeField.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
		HGrab:  true,
	})
	d.localeField.Tooltip = newWrappedTooltip(xstrings.Wrap("", i18n.Text(`The locale to use when presenting text in the user interface. This does not affect the content of data files. Leave this value blank to use the system default. Note that changes to this take effect the next time GCS is started.`), 100))
	d.localeField.Watermark = i18n.Locale()
	content.AddChild(d.localeField)
}

func (d *generalSettingsDockable) createDeepSearchCheckboxes(content *unison.Panel) {
	names := make(map[string]string)
	for _, ext := range gurps.DeepSearchableExtensions() {
		fi := gurps.FileInfoFor(ext)
		names[fi.UTI.Extensions[0]] = strings.TrimPrefix(fi.Name, "GCS ")
	}
	d.deepSearchableCheckbox = addMembershipCheckBoxPanel(content, i18n.Text("Library Explorer Deep Search"),
		slices.Collect(maps.Keys(names)), func(a, b string) int { return cmp.Compare(names[a], names[b]) },
		func(ext string) string { return names[ext] }, &gurps.GlobalSettings().DeepSearch,
		func() { Workspace.Navigator.mapDeepSearch() })
}

func (d *generalSettingsDockable) createOpenInWindowCheckboxes(content *unison.Panel) {
	d.openInWindowCheckbox = addMembershipCheckBoxPanel(content, i18n.Text("Use Separate Windows"), dgroup.Groups,
		cmp.Compare[dgroup.Group], dgroup.Group.String, &gurps.GlobalSettings().OpenInWindow, nil)
}

// addMembershipCheckBoxPanel adds a titled panel to the three-column content, spanning the field columns, holding a
// checkbox for each of the items, arranged in two columns in the order compare gives. Each checkbox toggles its item's
// membership in the sorted list that members points at; onChange, if not nil, runs after each such change. The pointer
// is stable because reset and load copy into the settings rather than replacing them.
func addMembershipCheckBoxPanel[T cmp.Ordered](content *unison.Panel, title string, items []T, compare func(a, b T) int,
	label func(T) string, members *[]T, onChange func(),
) []membershipCheckBox[T] {
	content.AddChild(unison.NewLabel())
	panel := unison.NewPanel()
	panel.SetBorder(unison.NewCompoundBorder(&TitledBorder{
		Title: title,
		Font:  unison.DefaultLabelTheme.Font,
	},
		unison.NewEmptyBorder(geom.NewSymmetricInsets(unison.StdHSpacing,
			unison.StdVSpacing))))
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	panel.SetLayoutData(&unison.FlexLayoutData{HSpan: 2})
	items = slices.Clone(items)
	xslices.ColumnSort(items, 2, compare)
	boxes := make([]membershipCheckBox[T], 0, len(items))
	for _, item := range items {
		box := NewCheckBox(nil, "", label(item),
			func() check.Enum {
				return check.FromBool(slices.Contains(*members, item))
			},
			func(state check.Enum) {
				i := slices.Index(*members, item)
				if state == check.On {
					if i == -1 {
						*members = append(*members, item)
						slices.Sort(*members)
					}
				} else if i != -1 {
					*members = slices.Delete(*members, i, i+1)
				}
				if onChange != nil {
					onChange()
				}
			})
		boxes = append(boxes, membershipCheckBox[T]{box: box, item: item})
		panel.AddChild(box)
	}
	content.AddChild(panel)
	return boxes
}

// syncMembership brings each checkbox into line with whether its item is among the members.
func syncMembership[T cmp.Ordered](boxes []membershipCheckBox[T], members []T) {
	for _, one := range boxes {
		SetCheckBoxState(one.box, slices.Contains(members, one.item))
	}
}

func (d *generalSettingsDockable) reset() {
	s := gurps.GlobalSettings()
	*s.General = *gurps.NewGeneralSettings()
	gurps.SyncScriptExecTimeLimit()
	s.DeepSearch = nil
	s.OpenInWindow = nil
	languageSetting = ""
	d.sync()
	Workspace.Navigator.mapDeepSearch()
}

func (d *generalSettingsDockable) sync() {
	s := gurps.GlobalSettings()
	gs := s.General
	d.nameField.SetText(gs.DefaultPlayerName)
	for _, one := range d.checkBoxes {
		SetCheckBoxState(one.box, *one.value)
	}
	d.appUpdateCheckPopup.Select(gs.AppUpdateCheck)
	d.libraryUpdateCheckPopup.Select(gs.LibraryUpdateCheck)
	d.pointsField.SetText(gs.InitialPoints.String())
	d.techLevelField.SetText(gs.DefaultTechLevel)
	d.calendarPopup.Select(gs.CalendarRef(s.Libraries).Name)
	for _, one := range d.initialScaleFields {
		SetFieldValue(one.field.Field, one.field.Format(*one.value))
	}
	d.autoScalingPopup.Select(gs.PDFAutoScaling)
	d.maxAutoColWidthField.SetText(strconv.Itoa(gs.MaximumAutoColWidth))
	d.monitorResolutionField.SetText(strconv.Itoa(gs.MonitorResolution))
	d.exportResolutionField.SetText(strconv.Itoa(gs.ImageResolution))
	d.permittedScriptExecTimeField.SetText(gs.PermittedPerScriptExecTime.String())
	d.tooltipDelayField.SetText(gs.TooltipDelay.String())
	d.tooltipDismissalField.SetText(gs.TooltipDismissal.String())
	d.scrollWheelMultiplierField.SetText(gs.ScrollWheelMultiplier.String())
	d.cursorSizeField.SetText(strconv.Itoa(gs.CursorSize))
	SetFieldValue(d.externalPDFCmdlineField.Field, gs.ExternalPDFCmdLine)
	SetFieldValue(d.localeField.Field, languageSetting)
	syncMembership(d.deepSearchableCheckbox, s.DeepSearch)
	syncMembership(d.openInWindowCheckbox, s.OpenInWindow)
	d.MarkForRedraw()
}

func (d *generalSettingsDockable) load(fileSystem fs.FS, filePath string) error {
	s, err := gurps.NewGeneralSettingsFromFile(fileSystem, filePath)
	if err != nil {
		return err
	}
	*gurps.GlobalSettings().General = *s
	gurps.SyncScriptExecTimeLimit()
	d.sync()
	return nil
}

func (d *generalSettingsDockable) save(filePath string) error {
	return gurps.GlobalSettings().General.Save(filePath)
}

// willClose persists the interface locale for the next launch. It deliberately does not apply the setting to
// i18n.Language: that global has no synchronization and is read by every i18n.Text call, including the ones the deep
// search content cache makes on its worker goroutines while parsing library files, so it may only be written before any
// of those goroutines exist, which is what LoadLanguageSetting does at launch.
func (d *generalSettingsDockable) willClose() bool {
	filePath := languageSettingPath()
	if languageSetting == "" {
		if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs.Log(err, "path", filePath)
		}
	} else if err := os.WriteFile(filePath, []byte(languageSetting), 0o640); err != nil {
		errs.Log(err, "path", filePath)
	}
	return true
}

// LoadLanguageSetting loads the language setting from disk, if present, and applies it. Must be called at launch,
// before any goroutine that may call i18n.Text is started, since i18n.Language is not safe to write once one exists.
func LoadLanguageSetting() {
	if data, err := os.ReadFile(languageSettingPath()); err == nil {
		if s := strings.TrimSpace(strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)[0]); len(s) > 1 && len(s) < 20 {
			i18n.Language = s
			languageSetting = s
		}
	}
}

func languageSettingPath() string {
	return filepath.Join(xos.AppDataDir(true), xos.AppCmdName+"_language.txt")
}
