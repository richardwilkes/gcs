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
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xflag"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xslices"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
)

// PathToLog is the path to the log file, set by the main entry point.
var PathToLog string

var languageSetting string

// generalSettingsMinSize is the smallest the General Settings view lets its content be shrunk to, which its widest
// row, the one pairing the initial image and PDF scales, needs to show whole.
var generalSettingsMinSize = geom.NewSize(440, 200)

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

// initialScaleField pairs an initial scale field with the general setting it edits, so that sync() can bring it back
// into line.
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
	if ActivateDockableOfType[*generalSettingsDockable](nil) {
		return
	}
	d := &generalSettingsDockable{}
	d.ContentWidthBehavior = behavior.Follow
	d.ContentMinSize = generalSettingsMinSize
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

// initContent builds the content as sections, each a grid of its own so that its label column is only as wide as its
// own labels, in the order the settings tend to be wanted, so that what a narrow view shows without scrolling is what
// is most often looked for.
func (d *generalSettingsDockable) initContent(content *unison.Panel) {
	initSettingsSections(content)
	d.createNewSheetsSection(content)
	d.createListsSection(content)
	d.createWorkspaceSection(content)
	d.createUpdatesSection(content)
	d.createInteractionSection(content)
	d.createAdvancedSection(content)
}

func (d *generalSettingsDockable) createNewSheetsSection(content *unison.Panel) {
	section := newSettingsSection(content, i18n.Text("New Sheets"), 4)
	gs := gurps.GlobalSettings().General
	name := i18n.Text("Default Player Name")
	d.nameField = NewStringField(nil, "", name,
		func() string { return gurps.GlobalSettings().General.DefaultPlayerName },
		func(s string) { gurps.GlobalSettings().General.DefaultPlayerName = s })
	d.nameField.Accessibility.Name = name
	addStretchedSettingsRow(section, i18n.Text("Player Name"), d.nameField, 3)
	name = i18n.Text("Initial Points")
	d.pointsField = NewDecimalField(nil, "", name,
		func() fxp.Int { return gurps.GlobalSettings().General.InitialPoints },
		func(v fxp.Int) { gurps.GlobalSettings().General.InitialPoints = v }, gurps.InitialPointsMin,
		gurps.InitialPointsMax, false, false)
	d.pointsField.Accessibility.Name = name
	addSettingsRow(section, i18n.Text("Points"), d.pointsField, 1)
	name = i18n.Text("Default Tech Level")
	d.techLevelField = NewStringField(nil, "", name,
		func() string { return gurps.GlobalSettings().General.DefaultTechLevel },
		func(s string) { gurps.GlobalSettings().General.DefaultTechLevel = s })
	d.techLevelField.Accessibility.Name = name
	d.techLevelField.Tooltip = newWrappedTooltip(gurps.TechLevelInfo())
	d.techLevelField.SetMinimumTextWidthUsing("12^")
	addSettingsRow(section, i18n.Text("Tech Level"), d.techLevelField, 1)
	d.createCalendarPopup(section)
	d.addGeneralCheckBox(section, i18n.Text("Fill in initial description"), &gs.AutoFillProfile, nil)
	d.addGeneralCheckBox(section, i18n.Text("Add natural attacks"), &gs.AutoAddNaturalAttacks, nil)
}

func (d *generalSettingsDockable) createCalendarPopup(section *unison.Panel) {
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
	addSettingsRow(section, i18n.Text("Calendar"), d.calendarPopup, 3)
}

func (d *generalSettingsDockable) createListsSection(content *unison.Panel) {
	section := newSettingsSection(content, i18n.Text("Lists"), 2)
	gs := gurps.GlobalSettings().General
	d.addGeneralCheckBox(section, i18n.Text("Group containers when sorting"), &gs.GroupContainersOnSort,
		func() { Workspace.Navigator.EventuallyReload() })
	box := d.addGeneralCheckBox(section, i18n.Text("Show additional page references when space allows"),
		&gs.ExpandPageReferences, func() {
			for _, one := range AllDockables() {
				if r, ok := one.(Rebuildable); ok {
					r.Rebuild(false)
				}
			}
		})
	box.Tooltip = newWrappedTooltip(i18n.Text(`When enabled, the page reference column on character sheets, loot sheets, and templates will display more than one page reference when there is room, rather than always collapsing to a single reference. Standalone lists are unaffected, since their columns can be resized directly.`))
	title := i18n.Text("Max Auto Column Width")
	d.maxAutoColWidthField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.MaximumAutoColWidth },
		func(v int) { gurps.GlobalSettings().General.MaximumAutoColWidth = v },
		gurps.AutoColWidthMin, gurps.AutoColWidthMax, false, false)
	addSettingsRow(section, title, d.maxAutoColWidthField, 1)
}

func (d *generalSettingsDockable) createWorkspaceSection(content *unison.Panel) {
	section := newSettingsSection(content, i18n.Text("Workspace"), 1)
	d.addGeneralCheckBox(section, i18n.Text("Restore arrangement on start"),
		&gurps.GlobalSettings().General.RestoreWorkspaceOnStart, nil)
	d.createOpenInWindowCheckboxes(section)
	d.createInitialScaleFields(section)
}

func (d *generalSettingsDockable) createUpdatesSection(content *unison.Panel) {
	section := newSettingsSection(content, i18n.Text("Updates & Library"), 2)
	d.createUpdateCheckPopups(section)
	d.createDeepSearchCheckboxes(section)
}

func (d *generalSettingsDockable) createInteractionSection(content *unison.Panel) {
	section := newSettingsSection(content, i18n.Text("Interaction"), 2)
	gs := gurps.GlobalSettings().General
	d.addGeneralCheckBox(section, i18n.Text("Initial click on text field selects all"),
		&gs.InitialFieldClickSelectsAll, nil)
	box := d.addGeneralCheckBox(section, i18n.Text("Tab stops on static text for screen readers"),
		&gs.FocusForReading, func() { gurps.GlobalSettings().General.UpdateFocusForReading() })
	box.Tooltip = newWrappedTooltip(i18n.Text(`Includes static text and disabled controls in the Tab order for screen readers. When enabled and a screen reader is in use, labels, other static text and disabled controls become Tab stops, as do the buttons a character sheet otherwise leaves out of the Tab order, such as its randomize buttons, so they can be reached by moving the keyboard focus alone. A disabled control reached this way is announced as unavailable and still cannot be changed. Nothing changes when no assistive technology is in use.`))
	d.createTooltipDelayField(section)
	d.createTooltipDismissalField(section)
	d.createScrollWheelMultiplierField(section)
	d.createCursorSizeField(section)
}

func (d *generalSettingsDockable) createAdvancedSection(content *unison.Panel) {
	section := newSettingsSection(content, i18n.Text("Advanced"), 3)
	d.createMonitorResolutionField(section)
	d.createImageResolutionField(section)
	d.createPermittedScriptExecTimeField(section)
	d.createExternalPDFCmdLineField(section)
	d.createLocaleField(section)
	d.createPathInfoField(section, i18n.Text("Settings Path"), i18n.Text("Copy the settings path to the clipboard"),
		gurps.SettingsPath)
	d.createPathInfoField(section, i18n.Text("Translations Path"),
		i18n.Text("Copy the translations path to the clipboard"), i18n.Dir)
	d.createPathInfoField(section, i18n.Text("Log Path"), i18n.Text("Copy the log path to the clipboard"), PathToLog)
}

// addGeneralCheckBox adds a checkbox for the general setting that value points at, spanning every column of the
// section. onChange, if not nil, runs after the setting has been changed through the checkbox or by a reset or load
// (see sync). The pointer is stable because reset and load copy into the settings rather than replacing them.
func (d *generalSettingsDockable) addGeneralCheckBox(section *unison.Panel, title string, value *bool, onChange func()) *CheckBox {
	box := addCheckBox(section, title, value)
	box.SetLayoutData(&unison.FlexLayoutData{HSpan: flexColumns(section)})
	box.OnSet = onChange
	d.checkBoxes = append(d.checkBoxes, settingCheckBox{box: box, value: value})
	return box
}

func (d *generalSettingsDockable) createUpdateCheckPopups(section *unison.Panel) {
	d.appUpdateCheckPopup = newUpdateCheckPopup(section,
		i18n.Text("Check for %s Updates", xos.AppName),
		i18n.Text("How often to look for a newer version of %s. Help ▸ Check for %s updates always works.",
			xos.AppName, xos.AppName),
		func() updatecheck.Option { return gurps.GlobalSettings().General.AppUpdateCheck },
		func(option updatecheck.Option) { gurps.GlobalSettings().General.AppUpdateCheck = option })
	d.libraryUpdateCheckPopup = newUpdateCheckPopup(section, i18n.Text("Check for Library Updates"),
		i18n.Text("How often to look for newer versions of the libraries. The Library Explorer's update buttons always work, checking for releases when clicked if no check has been made yet."),
		func() updatecheck.Option { return gurps.GlobalSettings().General.LibraryUpdateCheck },
		func(option updatecheck.Option) { gurps.GlobalSettings().General.LibraryUpdateCheck = option })
}

// newUpdateCheckPopup adds a labeled popup menu for choosing how often an update check should be made.
func newUpdateCheckPopup(section *unison.Panel, title, tooltip string, get func() updatecheck.Option,
	set func(updatecheck.Option),
) *unison.PopupMenu[updatecheck.Option] {
	popup := newPopupMenu(updatecheck.Options, get(), func(option updatecheck.Option) {
		set(option)
		ApplyUpdateCheckSettings()
	})
	popup.Tooltip = newWrappedTooltip(tooltip)
	addSettingsRow(section, title, popup, 1)
	return popup
}

func (d *generalSettingsDockable) createInitialScaleFields(section *unison.Panel) {
	gs := gurps.GlobalSettings().General
	grid := newSettingsSubgroup(section, i18n.Text("Initial scale"), 4)
	d.addInitialScaleField(grid, i18n.Text("List"), i18n.Text("Initial List Scale"), &gs.InitialListUIScale)
	d.addInitialScaleField(grid, i18n.Text("Editor"), i18n.Text("Initial Editor Scale"), &gs.InitialEditorUIScale)
	d.addInitialScaleField(grid, i18n.Text("Sheet"), i18n.Text("Initial Sheet Scale"), &gs.InitialSheetUIScale)
	d.addInitialScaleField(grid, i18n.Text("Markdown"), i18n.Text("Initial Markdown Scale"),
		&gs.InitialMarkdownUIScale)
	d.addInitialScaleField(grid, i18n.Text("Image"), i18n.Text("Initial Image Scale"), &gs.InitialImageUIScale)
	d.autoScalingPopup = newPopupMenu(autoscale.Options, gs.PDFAutoScaling,
		func(mode autoscale.Option) { gurps.GlobalSettings().General.PDFAutoScaling = mode })
	d.autoScalingPopup.Accessibility.Name = i18n.Text("PDF Auto Scaling")
	d.addInitialScaleField(grid, i18n.Text("PDF"), i18n.Text("Initial PDF Scale"), &gs.InitialPDFUIScale,
		d.autoScalingPopup)
}

// addInitialScaleField adds a percentage field for the initial scale setting that value points at, labeled with what
// the group's heading leaves to say but named outright with name, for screen readers, which hear no heading, followed
// on the same row by any trailing panels. The pointer is stable because reset and load copy into the settings rather
// than replacing them.
func (d *generalSettingsDockable) addInitialScaleField(grid *unison.Panel, label, name string, value *int, trailing ...unison.Paneler) *PercentageField {
	field := NewPercentageField(nil, "", name, func() int { return *value }, func(v int) { *value = v },
		gurps.InitialUIScaleMin, gurps.InitialUIScaleMax, false, false)
	field.Accessibility.Name = name
	addSettingsRow(grid, label, field, 1, trailing...)
	d.initialScaleFields = append(d.initialScaleFields, initialScaleField{field: field, value: value})
	return field
}

func (d *generalSettingsDockable) createMonitorResolutionField(section *unison.Panel) {
	title := i18n.Text("Monitor Resolution")
	d.monitorResolutionField = NewNumericFieldWithException(nil, "", title,
		func(minValue, maxValue int) []int { return []int{minValue, maxValue} },
		func() int { return gurps.GlobalSettings().General.MonitorResolution },
		func(v int) { gurps.GlobalSettings().General.MonitorResolution = v },
		strconv.Itoa, strconv.Atoi, gurps.MonitorResolutionMin, gurps.MonitorResolutionMax, 0)
	d.monitorResolutionField.Tooltip = newWrappedTooltip(i18n.Text("A value of 0 uses the ppi your monitor reports."))
	addSettingsRow(section, title, d.monitorResolutionField, 2,
		NewFieldTrailingHint(d.monitorResolutionField, i18n.Text("ppi"), false))
}

func (d *generalSettingsDockable) createImageResolutionField(section *unison.Panel) {
	title := i18n.Text("Image Export Resolution")
	d.exportResolutionField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.ImageResolution },
		func(v int) { gurps.GlobalSettings().General.ImageResolution = v },
		gurps.ImageResolutionMin, gurps.ImageResolutionMax, false, false)
	addSettingsRow(section, title, d.exportResolutionField, 2,
		NewFieldTrailingHint(d.exportResolutionField, i18n.Text("ppi"), false))
}

func (d *generalSettingsDockable) createPermittedScriptExecTimeField(section *unison.Panel) {
	title := i18n.Text("Script Execution Limit")
	d.permittedScriptExecTimeField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.PermittedPerScriptExecTime },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.PermittedPerScriptExecTime = v
			gurps.SyncScriptExecTimeLimit()
		}, gurps.PermittedScriptExecTimeMin, gurps.PermittedScriptExecTimeMax, false, false)
	addSettingsRow(section, title, d.permittedScriptExecTimeField, 2,
		NewFieldTrailingHint(d.permittedScriptExecTimeField, i18n.Text("seconds per script"), false))
}

func (d *generalSettingsDockable) createTooltipDelayField(section *unison.Panel) {
	title := i18n.Text("Tooltip Delay")
	d.tooltipDelayField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.TooltipDelay },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.TooltipDelay = v
			general.UpdateToolTipTiming()
		}, gurps.TooltipDelayMin, gurps.TooltipDelayMax, false, false)
	addSettingsRow(section, title, d.tooltipDelayField, 1,
		NewFieldTrailingHint(d.tooltipDelayField, i18n.Text("seconds"), false))
}

func (d *generalSettingsDockable) createTooltipDismissalField(section *unison.Panel) {
	title := i18n.Text("Tooltip Dismissal")
	d.tooltipDismissalField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.TooltipDismissal },
		func(v fxp.Int) {
			general := gurps.GlobalSettings().General
			general.TooltipDismissal = v
			general.UpdateToolTipTiming()
		}, gurps.TooltipDismissalMin, gurps.TooltipDismissalMax, false, false)
	addSettingsRow(section, title, d.tooltipDismissalField, 1,
		NewFieldTrailingHint(d.tooltipDismissalField, i18n.Text("seconds"), false))
}

func (d *generalSettingsDockable) createScrollWheelMultiplierField(section *unison.Panel) {
	title := i18n.Text("Scroll Wheel Multiplier")
	d.scrollWheelMultiplierField = NewDecimalField(nil, "", title,
		func() fxp.Int { return gurps.GlobalSettings().General.ScrollWheelMultiplier },
		func(v fxp.Int) { gurps.GlobalSettings().General.ScrollWheelMultiplier = v },
		gurps.ScrollWheelMultiplierMin, gurps.ScrollWheelMultiplierMax, false, false)
	addSettingsRow(section, title, d.scrollWheelMultiplierField, 1)
}

func (d *generalSettingsDockable) createCursorSizeField(section *unison.Panel) {
	title := i18n.Text("Cursor Size")
	d.cursorSizeField = NewIntegerField(nil, "", title,
		func() int { return gurps.GlobalSettings().General.CursorSize },
		func(v int) {
			general := gurps.GlobalSettings().General
			general.CursorSize = v
			general.UpdateCursorSize()
		}, gurps.CursorSizeMin, gurps.CursorSizeMax, false, false)
	addSettingsRow(section, title, d.cursorSizeField, 1,
		NewFieldTrailingHint(d.cursorSizeField, i18n.Text("points"), false))
}

// createPathInfoField adds a row showing the path, with a button that copies it to the clipboard. The button is named
// by copyTooltip, which has to say which path it copies, since the buttons of the rows would otherwise sound alike.
func (d *generalSettingsDockable) createPathInfoField(section *unison.Panel, title, copyTooltip, value string) {
	section.AddChild(NewFieldLeadingLabel(title, false))
	section.AddChild(NewPathField(value))
	copyButton := unison.NewSVGButton(svg.Copy)
	copyButton.Tooltip = newWrappedTooltip(copyTooltip)
	copyButton.ClickCallback = func() {
		unison.ClipboardSetText(value)
	}
	section.AddChild(copyButton)
}

func (d *generalSettingsDockable) createExternalPDFCmdLineField(section *unison.Panel) {
	title := i18n.Text("External PDF Viewer")
	d.externalPDFCmdlineField = NewStringField(nil, "", title,
		func() string { return gurps.GlobalSettings().General.ExternalPDFCmdLine },
		func(s string) { gurps.GlobalSettings().General.ExternalPDFCmdLine = strings.TrimSpace(s) })
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
	addStretchedSettingsRow(section, title, d.externalPDFCmdlineField, 2)
}

func (d *generalSettingsDockable) createLocaleField(section *unison.Panel) {
	title := i18n.Text("Interface Locale")
	d.localeField = NewStringField(nil, "", title,
		func() string { return languageSetting },
		func(s string) { languageSetting = strings.TrimSpace(s) })
	d.localeField.Tooltip = newWrappedTooltip(i18n.Text(`The locale to use when presenting text in the user interface. This does not affect the content of data files. Leave this value blank to use the system default. Note that changes to this take effect the next time GCS is started.`))
	d.localeField.Watermark = i18n.Locale()
	addStretchedSettingsRow(section, title, d.localeField, 2)
}

func (d *generalSettingsDockable) createDeepSearchCheckboxes(section *unison.Panel) {
	names := make(map[string]string)
	for _, ext := range gurps.DeepSearchableExtensions() {
		fi := gurps.FileInfoFor(ext)
		names[fi.UTI.Extensions[0]] = strings.TrimPrefix(fi.Name, "GCS ")
	}
	d.deepSearchableCheckbox = addMembershipCheckBoxPanel(section, i18n.Text("Deep search"),
		slices.Collect(maps.Keys(names)), func(a, b string) int { return cmp.Compare(names[a], names[b]) },
		func(ext string) string { return names[ext] }, &gurps.GlobalSettings().DeepSearch,
		func() { Workspace.Navigator.mapDeepSearch() })
}

func (d *generalSettingsDockable) createOpenInWindowCheckboxes(section *unison.Panel) {
	d.openInWindowCheckbox = addMembershipCheckBoxPanel(section, i18n.Text("Open in separate windows"),
		dgroup.Groups, cmp.Compare[dgroup.Group], dgroup.Group.String, &gurps.GlobalSettings().OpenInWindow, nil)
}

// addMembershipCheckBoxPanel adds a group with the title to the section, holding a checkbox for each of the items,
// arranged in two columns in the order compare gives. Each checkbox toggles its item's membership in the sorted list
// that members points at; onChange, if not nil, runs after each such change. The pointer is stable because reset and
// load copy into the settings rather than replacing them.
func addMembershipCheckBoxPanel[T cmp.Ordered](section *unison.Panel, title string, items []T, compare func(a, b T) int,
	label func(T) string, members *[]T, onChange func(),
) []membershipCheckBox[T] {
	grid := newSettingsSubgroup(section, title, 2)
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
		grid.AddChild(box)
	}
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
	s.General.EnsureValidity() // Pushes the defaults unison keeps copies of, such as tooltip timing and cursor size.
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
		// Sync runs the box's OnSet when a reset or load changed its setting, as a click would.
		one.box.Sync()
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
