// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"io/fs"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/autoscale"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/unison"
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
	ScrollWheelMultiplierDef   = fxp.From(unison.MouseWheelMultiplier)
	ScrollWheelMultiplierMin   = fxp.One
	ScrollWheelMultiplierMax   = fxp.TenThousandMinusOne
	PermittedScriptExecTimeDef = fxp.FromStringForced("0.05")
	PermittedScriptExecTimeMin = fxp.FromStringForced("0.001")
	PermittedScriptExecTimeMax = fxp.Half
)

// Default, minimum & maximum values for the general numeric settings that can be constants
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
	InitialSheetUIScaleDef     = 150
	InitialPDFUIScaleDef       = 100
	InitialMarkdownUIScaleDef  = 100
	InitialImageUIScaleDef     = 100
	InitialPDFAutoScaling      = autoscale.No
	AutoColWidthMin            = 50
	AutoColWidthMax            = 9999
	MaximumAutoColWidthDef     = 800
)

// GeneralSettings holds general settings for a sheet.
type GeneralSettings struct {
	DefaultPlayerName           string           `json:"default_player_name,omitempty"`
	DefaultTechLevel            string           `json:"default_tech_level,omitempty"`
	CalendarName                string           `json:"calendar_ref,omitempty"`
	ExternalPDFCmdLine          string           `json:"external_pdf_cmd_line,omitempty"`
	InitialPoints               fxp.Int          `json:"initial_points"`
	TooltipDelay                fxp.Int          `json:"tooltip_delay"`
	TooltipDismissal            fxp.Int          `json:"tooltip_dismissal"`
	ScrollWheelMultiplier       fxp.Int          `json:"scroll_wheel_multiplier"`
	PermittedPerScriptExecTime  fxp.Int          `json:"permitted_per_script_exec_time,omitempty"`
	NavigatorUIScale            int              `json:"navigator_scale"`
	InitialListUIScale          int              `json:"initial_list_scale"`
	InitialEditorUIScale        int              `json:"initial_editor_scale"`
	InitialSheetUIScale         int              `json:"initial_sheet_scale"`
	InitialPDFUIScale           int              `json:"initial_pdf_scale"`
	InitialMarkdownUIScale      int              `json:"initial_md_scale"`
	InitialImageUIScale         int              `json:"initial_img_scale"`
	MaximumAutoColWidth         int              `json:"maximum_auto_col_width"`
	ImageResolution             int              `json:"image_resolution"`
	MonitorResolution           int              `json:"monitor_resolution,omitempty"`
	PDFAutoScaling              autoscale.Option `json:"pdf_auto_scaling,omitempty"`
	AutoFillProfile             bool             `json:"auto_fill_profile"`
	AutoAddNaturalAttacks       bool             `json:"add_natural_attacks"`
	GroupContainersOnSort       bool             `json:"group_containers_on_sort"`
	InitialFieldClickSelectsAll bool             `json:"initial_field_click_selects_all"`
	RestoreWorkspaceOnStart     bool             `json:"restore_workspace_on_start"`
}

// NewGeneralSettings creates settings with factory defaults.
func NewGeneralSettings() *GeneralSettings {
	return &GeneralSettings{
		DefaultPlayerName:          toolbox.CurrentUserName(),
		DefaultTechLevel:           "3",
		InitialPoints:              InitialPointsDef,
		TooltipDelay:               TooltipDelayDef,
		TooltipDismissal:           TooltipDismissalDef,
		ScrollWheelMultiplier:      fxp.From(unison.MouseWheelMultiplier),
		PermittedPerScriptExecTime: PermittedScriptExecTimeDef,
		NavigatorUIScale:           InitialNavigatorUIScaleDef,
		InitialListUIScale:         InitialListUIScaleDef,
		InitialEditorUIScale:       InitialEditorUIScaleDef,
		InitialSheetUIScale:        InitialSheetUIScaleDef,
		InitialPDFUIScale:          InitialPDFUIScaleDef,
		InitialMarkdownUIScale:     InitialMarkdownUIScaleDef,
		InitialImageUIScale:        InitialImageUIScaleDef,
		MaximumAutoColWidth:        MaximumAutoColWidthDef,
		ImageResolution:            ImageResolutionDef,
		PDFAutoScaling:             InitialPDFAutoScaling,
		AutoFillProfile:            true,
		AutoAddNaturalAttacks:      true,
		RestoreWorkspaceOnStart:    true,
	}
}

// NewGeneralSettingsFromFile loads new settings from a file.
func NewGeneralSettingsFromFile(fileSystem fs.FS, filePath string) (*GeneralSettings, error) {
	var data struct {
		GeneralSettings
		OldLocation *GeneralSettings `json:"general"`
	}
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &data); err != nil {
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
	return jio.SaveToFile(context.Background(), filePath, s)
}

// UpdateToolTipTiming updates the default tooltip theme to use the timing values from this object.
func (s *GeneralSettings) UpdateToolTipTiming() {
	unison.DefaultTooltipTheme.Delay = fxp.SecondsToDuration(s.TooltipDelay)
	unison.DefaultTooltipTheme.Dismissal = fxp.SecondsToDuration(s.TooltipDismissal)
}

// CalendarRef returns the CalendarRef these settings refer to.
func (s *GeneralSettings) CalendarRef(libraries Libraries) *CalendarRef {
	ref := LookupCalendarRef(s.CalendarName, libraries)
	if ref == nil {
		if ref = LookupCalendarRef("Gregorian", libraries); ref == nil {
			fatal.IfErr(errs.New("unable to load default calendar (Gregorian)"))
		}
	}
	return ref
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (s *GeneralSettings) EnsureValidity() {
	s.InitialPoints = fxp.ResetIfOutOfRange(s.InitialPoints, InitialPointsMin, InitialPointsMax, InitialPointsDef)
	s.TooltipDelay = fxp.ResetIfOutOfRange(s.TooltipDelay, TooltipDelayMin, TooltipDelayMax, TooltipDelayDef)
	s.TooltipDismissal = fxp.ResetIfOutOfRange(s.TooltipDismissal, TooltipDismissalMin, TooltipDismissalMax,
		TooltipDismissalDef)
	s.PermittedPerScriptExecTime = fxp.ResetIfOutOfRange(s.PermittedPerScriptExecTime, PermittedScriptExecTimeMin,
		PermittedScriptExecTimeMax, PermittedScriptExecTimeDef)
	s.ScrollWheelMultiplier = fxp.ResetIfOutOfRange(s.ScrollWheelMultiplier, ScrollWheelMultiplierMin,
		ScrollWheelMultiplierMax, ScrollWheelMultiplierDef)
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
	s.PDFAutoScaling = s.PDFAutoScaling.EnsureValid()
	s.UpdateToolTipTiming()
}
