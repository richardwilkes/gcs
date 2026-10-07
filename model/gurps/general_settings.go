// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/autoscale"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/updatecheck"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
)

// Default, minimum & maximum values for the general numeric settings
var (
	InitialPointsDef           = fxp.OneHundredFifty
	InitialPointsMin           fxp.Int
	InitialPointsMax           = fxp.TenMillionMinusOne
	TooltipDelayDef            = fxp.ThreeQuarters
	TooltipDelayMin            fxp.Int
	TooltipDelayMax            = fxp.Thirty
	TooltipDismissalDef        = fxp.Sixty
	TooltipDismissalMin        = fxp.One
	TooltipDismissalMax        = fxp.ThirtySixHundred
	ScrollWheelMultiplierMin   = fxp.One
	ScrollWheelMultiplierMax   = fxp.TenThousandMinusOne
	PermittedScriptExecTimeDef = fxp.FromStringForced("0.05")
	PermittedScriptExecTimeMin = fxp.FromStringForced("0.001")
	PermittedScriptExecTimeMax = fxp.Half
)

// Default, minimum & maximum values for the general numeric settings that can be constants. The cursor sizes match
// those of the user interface, which a test there verifies.
const (
	MonitorResolutionMin       = 72
	MonitorResolutionMax       = 300
	ImageResolutionDef         = 200
	ImageResolutionMin         = 50
	ImageResolutionMax         = 400
	InitialUIScaleMin          = 50
	InitialUIScaleMax          = 400
	InitialNavigatorUIScaleDef = 100
	InitialListUIScaleDef      = 100
	InitialEditorUIScaleDef    = 100
	InitialSheetUIScaleDef     = 100
	InitialPDFUIScaleDef       = 100
	InitialMarkdownUIScaleDef  = 100
	InitialImageUIScaleDef     = 100
	InitialPDFAutoScaling      = autoscale.No
	InitialAppUpdateCheck      = updatecheck.AtLaunch
	InitialLibraryUpdateCheck  = updatecheck.AtLaunch
	AutoColWidthMin            = 50
	AutoColWidthMax            = 9999
	MaximumAutoColWidthDef     = 800
	CursorSizeDef              = 24
	CursorSizeMin              = 8
	CursorSizeMax              = 128
)

// defaultMonitorPPI is the monitor PPI used when neither the settings nor the host supply one.
const defaultMonitorPPI = 108

const currentGeneralSettingsVersion = 2

// GeneralSettings holds the application's general settings.
type GeneralSettings struct {
	DefaultPlayerName           string             `json:"default_player_name,omitzero"`
	DefaultTechLevel            string             `json:"default_tech_level,omitzero"`
	CalendarName                string             `json:"calendar_ref,omitzero"`
	ExternalPDFCmdLine          string             `json:"external_pdf_cmd_line,omitzero"`
	InitialPoints               fxp.Int            `json:"initial_points"`
	TooltipDelay                fxp.Int            `json:"tooltip_delay"`
	TooltipDismissal            fxp.Int            `json:"tooltip_dismissal"`
	ScrollWheelMultiplier       fxp.Int            `json:"scroll_wheel_multiplier"`
	PermittedPerScriptExecTime  fxp.Int            `json:"permitted_per_script_exec_time,omitzero"`
	Version                     int                `json:"version,omitzero"`
	NavigatorUIScale            int                `json:"navigator_scale"`
	InitialListUIScale          int                `json:"initial_list_scale"`
	InitialEditorUIScale        int                `json:"initial_editor_scale"`
	InitialSheetUIScale         int                `json:"initial_sheet_scale"`
	InitialPDFUIScale           int                `json:"initial_pdf_scale"`
	InitialMarkdownUIScale      int                `json:"initial_md_scale"`
	InitialImageUIScale         int                `json:"initial_img_scale"`
	MaximumAutoColWidth         int                `json:"maximum_auto_col_width"`
	ImageResolution             int                `json:"image_resolution"`
	MonitorResolution           int                `json:"monitor_resolution,omitzero"`
	CursorSize                  int                `json:"cursor_size"`
	PDFAutoScaling              autoscale.Option   `json:"pdf_auto_scaling,omitzero"`
	AppUpdateCheck              updatecheck.Option `json:"app_update_check,omitzero"`
	LibraryUpdateCheck          updatecheck.Option `json:"library_update_check,omitzero"`
	AutoFillProfile             bool               `json:"auto_fill_profile"`
	AutoAddNaturalAttacks       bool               `json:"add_natural_attacks"`
	GroupContainersOnSort       bool               `json:"group_containers_on_sort"`
	InitialFieldClickSelectsAll bool               `json:"initial_field_click_selects_all"`
	RestoreWorkspaceOnStart     bool               `json:"restore_workspace_on_start"`
	ExpandPageReferences        bool               `json:"expand_page_references"`
	FocusForReading             bool               `json:"focus_for_reading,omitzero"`
}

// NewGeneralSettings creates settings with factory defaults.
func NewGeneralSettings() *GeneralSettings {
	return &GeneralSettings{
		DefaultPlayerName:          xos.CurrentUserName(),
		DefaultTechLevel:           "3",
		InitialPoints:              InitialPointsDef,
		TooltipDelay:               TooltipDelayDef,
		TooltipDismissal:           TooltipDismissalDef,
		ScrollWheelMultiplier:      fxp.FromFloat(currentHost().DefaultScrollWheelMultiplier()),
		PermittedPerScriptExecTime: PermittedScriptExecTimeDef,
		Version:                    currentGeneralSettingsVersion,
		NavigatorUIScale:           InitialNavigatorUIScaleDef,
		InitialListUIScale:         InitialListUIScaleDef,
		InitialEditorUIScale:       InitialEditorUIScaleDef,
		InitialSheetUIScale:        InitialSheetUIScaleDef,
		InitialPDFUIScale:          InitialPDFUIScaleDef,
		InitialMarkdownUIScale:     InitialMarkdownUIScaleDef,
		InitialImageUIScale:        InitialImageUIScaleDef,
		MaximumAutoColWidth:        MaximumAutoColWidthDef,
		ImageResolution:            ImageResolutionDef,
		CursorSize:                 CursorSizeDef,
		PDFAutoScaling:             InitialPDFAutoScaling,
		AppUpdateCheck:             InitialAppUpdateCheck,
		LibraryUpdateCheck:         InitialLibraryUpdateCheck,
		AutoFillProfile:            true,
		AutoAddNaturalAttacks:      true,
		RestoreWorkspaceOnStart:    true,
		ExpandPageReferences:       true,
	}
}

// NewGeneralSettingsFromFile loads new settings from a file.
func NewGeneralSettingsFromFile(fileSystem fs.FS, filePath string) (*GeneralSettings, error) {
	var data struct {
		GeneralSettings
		OldLocation *GeneralSettings `json:"general"`
	}
	if err := jio.LoadFromFile(fileSystem, filePath, &data); err != nil {
		return nil, err
	}
	var s *GeneralSettings
	if data.OldLocation != nil {
		s = data.OldLocation
	} else {
		settings := data.GeneralSettings
		s = &settings
	}
	s.EnsureValidity()
	return s, nil
}

// Save writes the settings to the file as JSON.
func (s *GeneralSettings) Save(filePath string) error {
	s.EnsureValidity()
	return jio.SaveToFile(filePath, s)
}

// UpdateToolTipTiming hands the tooltip timing values from this object to the host.
func (s *GeneralSettings) UpdateToolTipTiming() {
	currentHost().SetTooltipTiming(fxp.SecondsToDuration(s.TooltipDelay), fxp.SecondsToDuration(s.TooltipDismissal))
}

// UpdateCursorSize hands the cursor size from this object to the host.
func (s *GeneralSettings) UpdateCursorSize() {
	currentHost().SetCursorSize(s.CursorSize)
}

// UpdateFocusForReading hands the FocusForReading setting to the host, which keeps its own copy.
func (s *GeneralSettings) UpdateFocusForReading() {
	currentHost().SetFocusForReading(s.FocusForReading)
}

// applyToHost hands the host every setting it keeps a copy of, which the Update methods do one at a time for the
// current host.
func (s *GeneralSettings) applyToHost(h Host) {
	h.SetTooltipTiming(fxp.SecondsToDuration(s.TooltipDelay), fxp.SecondsToDuration(s.TooltipDismissal))
	h.SetCursorSize(s.CursorSize)
	h.SetFocusForReading(s.FocusForReading)
}

// CalendarRef returns the CalendarRef these settings refer to.
func (s *GeneralSettings) CalendarRef(libraries *library.Libraries) *CalendarRef {
	ref := LookupCalendarRef(s.CalendarName, libraries)
	if ref == nil {
		if ref = LookupCalendarRef("Gregorian", libraries); ref == nil {
			xos.ExitIfErr(errs.New("unable to load default calendar (Gregorian)"))
		}
	}
	return ref
}

// MonitorPPI returns the monitor PPI to use, either from the settings or from the primary display, defaulting to
// defaultMonitorPPI when the host has no usable value for the display. It asks the host for the display, so it may
// only be called on the UI thread.
func (s *GeneralSettings) MonitorPPI() int {
	if s.MonitorResolution != 0 {
		return s.MonitorResolution
	}
	if ppi := currentHost().PrimaryDisplayPPI(); ppi > 0 {
		return ppi
	}
	return defaultMonitorPPI
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (s *GeneralSettings) EnsureValidity() {
	if s.Version != currentGeneralSettingsVersion {
		if s.Version < 1 {
			// Adjust from old default of 150% to new default of 100%
			s.InitialSheetUIScale = s.InitialSheetUIScale * 100 / 150
		}
		if s.Version < 2 {
			// This setting didn't exist before, so enable it to match the new default behavior.
			s.ExpandPageReferences = true
		}
		s.Version = currentGeneralSettingsVersion
	}
	s.InitialPoints = fxp.ResetIfOutOfRange(s.InitialPoints, InitialPointsMin, InitialPointsMax, InitialPointsDef)
	s.TooltipDelay = fxp.ResetIfOutOfRange(s.TooltipDelay, TooltipDelayMin, TooltipDelayMax, TooltipDelayDef)
	s.TooltipDismissal = fxp.ResetIfOutOfRange(s.TooltipDismissal, TooltipDismissalMin, TooltipDismissalMax,
		TooltipDismissalDef)
	s.PermittedPerScriptExecTime = fxp.ResetIfOutOfRange(s.PermittedPerScriptExecTime, PermittedScriptExecTimeMin,
		PermittedScriptExecTimeMax, PermittedScriptExecTimeDef)
	s.ScrollWheelMultiplier = fxp.ResetIfOutOfRange(s.ScrollWheelMultiplier, ScrollWheelMultiplierMin,
		ScrollWheelMultiplierMax, fxp.FromFloat(currentHost().DefaultScrollWheelMultiplier()))
	if s.MonitorResolution != 0 {
		s.MonitorResolution = fxp.ResetIfOutOfRange(s.MonitorResolution, MonitorResolutionMin, MonitorResolutionMax, 0)
	}
	s.ImageResolution = fxp.ResetIfOutOfRange(s.ImageResolution, ImageResolutionMin, ImageResolutionMax,
		ImageResolutionDef)
	s.NavigatorUIScale = fxp.ResetIfOutOfRange(s.NavigatorUIScale, InitialUIScaleMin, InitialUIScaleMax,
		InitialNavigatorUIScaleDef)
	s.InitialListUIScale = fxp.ResetIfOutOfRange(s.InitialListUIScale, InitialUIScaleMin, InitialUIScaleMax,
		InitialListUIScaleDef)
	s.InitialEditorUIScale = fxp.ResetIfOutOfRange(s.InitialEditorUIScale, InitialUIScaleMin, InitialUIScaleMax,
		InitialEditorUIScaleDef)
	s.InitialSheetUIScale = fxp.ResetIfOutOfRange(s.InitialSheetUIScale, InitialUIScaleMin, InitialUIScaleMax,
		InitialSheetUIScaleDef)
	s.InitialPDFUIScale = fxp.ResetIfOutOfRange(s.InitialPDFUIScale, InitialUIScaleMin, InitialUIScaleMax,
		InitialPDFUIScaleDef)
	s.InitialMarkdownUIScale = fxp.ResetIfOutOfRange(s.InitialMarkdownUIScale, InitialUIScaleMin, InitialUIScaleMax,
		InitialMarkdownUIScaleDef)
	s.InitialImageUIScale = fxp.ResetIfOutOfRange(s.InitialImageUIScale, InitialUIScaleMin, InitialUIScaleMax,
		InitialImageUIScaleDef)
	s.MaximumAutoColWidth = fxp.ResetIfOutOfRange(s.MaximumAutoColWidth, AutoColWidthMin, AutoColWidthMax,
		MaximumAutoColWidthDef)
	s.CursorSize = fxp.ResetIfOutOfRange(s.CursorSize, CursorSizeMin, CursorSizeMax, CursorSizeDef)
	s.PDFAutoScaling = s.PDFAutoScaling.EnsureValid()
	s.AppUpdateCheck = s.AppUpdateCheck.EnsureValid()
	s.LibraryUpdateCheck = s.LibraryUpdateCheck.EnsureValid()
	s.applyToHost(currentHost())
}
