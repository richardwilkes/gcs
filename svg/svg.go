// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package svg

import (
	_ "embed"
	"image"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/unison"
)

// Pre-defined SVG images used by GCS.
var (
	//go:embed attributes.svg
	attributesData string
	Attributes     = unison.MustSVGFromContentString(attributesData)

	//go:embed back.svg
	backData string
	Back     = unison.MustSVGFromContentString(backData)

	//go:embed body_type.svg
	bodyTypeData string
	BodyType     = unison.MustSVGFromContentString(bodyTypeData)

	//go:embed bookmark.svg
	bookmarkData string
	Bookmark     = unison.MustSVGFromContentString(bookmarkData)

	//go:embed calculator.svg
	calculatorData string
	Calculator     = unison.MustSVGFromContentString(calculatorData)

	//go:embed circled_add.svg
	circledAddData string
	CircledAdd     = unison.MustSVGFromContentString(circledAddData)

	//go:embed circled_vertical_ellipsis.svg
	circledVerticalEllipsisData string
	CircledVerticalEllipsis     = unison.MustSVGFromContentString(circledVerticalEllipsisData)

	//go:embed clone.svg
	cloneData string
	Clone     = unison.MustSVGFromContentString(cloneData)

	//go:embed closed_folder.svg
	closedFolderData string
	ClosedFolder     = unison.MustSVGFromContentString(closedFolderData)

	//go:embed coins.svg
	coinsData string
	Coins     = unison.MustSVGFromContentString(coinsData)

	//go:embed copy.svg
	copyData string
	Copy     = unison.MustSVGFromContentString(copyData)

	//go:embed database.svg
	databaseData string
	Database     = unison.MustSVGFromContentString(databaseData)

	//go:embed down_to_bracket.svg
	downToBracketData string
	DownToBracket     = unison.MustSVGFromContentString(downToBracketData)

	//go:embed download.svg
	downloadData string
	Download     = unison.MustSVGFromContentString(downloadData)

	//go:embed edit.svg
	editData string
	Edit     = unison.MustSVGFromContentString(editData)

	//go:embed first.svg
	firstData string
	First     = unison.MustSVGFromContentString(firstData)

	//go:embed first_aid_kit.svg
	firstAidKitData string
	FirstAidKit     = unison.MustSVGFromContentString(firstAidKitData)

	//go:embed forward.svg
	forwardData string
	Forward     = unison.MustSVGFromContentString(forwardData)

	//go:embed gcs_campaign.svg
	gcsCampaignData string
	GCSCampaign     = unison.MustSVGFromContentString(gcsCampaignData)

	//go:embed gcs_equipment.svg
	gcsEquipmentData string
	GCSEquipment     = unison.MustSVGFromContentString(gcsEquipmentData)

	//go:embed gcs_equipment_modifiers.svg
	gcsEquipmentModifiersData string
	GCSEquipmentModifiers     = unison.MustSVGFromContentString(gcsEquipmentModifiersData)

	//go:embed gcs_loot.svg
	gcsLootData string
	GCSLoot     = unison.MustSVGFromContentString(gcsLootData)

	//go:embed gcs_notes.svg
	gcsNotesData string
	GCSNotes     = unison.MustSVGFromContentString(gcsNotesData)

	//go:embed gcs_sheet.svg
	gcsSheetData string
	GCSSheet     = unison.MustSVGFromContentString(gcsSheetData)

	//go:embed gcs_skills.svg
	gcsSkillsData string
	GCSSkills     = unison.MustSVGFromContentString(gcsSkillsData)

	//go:embed gcs_spells.svg
	gcsSpellsData string
	GCSSpells     = unison.MustSVGFromContentString(gcsSpellsData)

	//go:embed gcs_template.svg
	gcsTemplateData string
	GCSTemplate     = unison.MustSVGFromContentString(gcsTemplateData)

	//go:embed gcs_trait_modifiers.svg
	gcsTraitModifiersData string
	GCSTraitModifiers     = unison.MustSVGFromContentString(gcsTraitModifiersData)

	//go:embed gcs_traits.svg
	gcsTraitsData string
	GCSTraits     = unison.MustSVGFromContentString(gcsTraitsData)

	//go:embed gears.svg
	gearsData string
	Gears     = unison.MustSVGFromContentString(gearsData)

	//go:embed generic_file.svg
	genericFileData string
	GenericFile     = unison.MustSVGFromContentString(genericFileData)

	//go:embed grip.svg
	gripData string
	Grip     = unison.MustSVGFromContentString(gripData)

	//go:embed help.svg
	helpData string
	Help     = unison.MustSVGFromContentString(helpData)

	//go:embed hierarchy.svg
	hierarchyData string
	Hierarchy     = unison.MustSVGFromContentString(hierarchyData)

	//go:embed image_file.svg
	imageFileData string
	ImageFile     = unison.MustSVGFromContentString(imageFileData)

	//go:embed info.svg
	infoData string
	Info     = unison.MustSVGFromContentString(infoData)

	//go:embed last.svg
	lastData string
	Last     = unison.MustSVGFromContentString(lastData)

	//go:embed link.svg
	linkData string
	Link     = unison.MustSVGFromContentString(linkData)

	//go:embed markdown_file.svg
	markdownFileData string
	MarkdownFile     = unison.MustSVGFromContentString(markdownFileData)

	//go:embed melee_weapon.svg
	meleeWeaponData string
	MeleeWeapon     = unison.MustSVGFromContentString(meleeWeaponData)

	//go:embed menu.svg
	menuData string
	Menu     = unison.MustSVGFromContentString(menuData)

	//go:embed naming.svg
	namingData string
	Naming     = unison.MustSVGFromContentString(namingData)

	//go:embed new_folder.svg
	newFolderData string
	NewFolder     = unison.MustSVGFromContentString(newFolderData)

	//go:embed next.svg
	nextData string
	Next     = unison.MustSVGFromContentString(nextData)

	//go:embed not.svg
	notData string
	Not     = unison.MustSVGFromContentString(notData)

	//go:embed notes-collapse.svg
	notesCollapseData string
	NotesCollapse     = unison.MustSVGFromContentString(notesCollapseData)

	//go:embed notes-expand.svg
	notesExpandData string
	NotesExpand     = unison.MustSVGFromContentString(notesExpandData)

	//go:embed notes-toggle.svg
	notesToggleData string
	NotesToggle     = unison.MustSVGFromContentString(notesToggleData)

	//go:embed open_folder.svg
	openFolderData string
	OpenFolder     = unison.MustSVGFromContentString(openFolderData)

	//go:embed pdf_file.svg
	pdfFileData string
	PDFFile     = unison.MustSVGFromContentString(pdfFileData)

	//go:embed previous.svg
	previousData string
	Previous     = unison.MustSVGFromContentString(previousData)

	//go:embed randomize.svg
	randomizeData string
	Randomize     = unison.MustSVGFromContentString(randomizeData)

	//go:embed ranged_weapon.svg
	rangedWeaponData string
	RangedWeapon     = unison.MustSVGFromContentString(rangedWeaponData)

	//go:embed release_notes.svg
	releaseNotesData string
	ReleaseNotes     = unison.MustSVGFromContentString(releaseNotesData)

	//go:embed reset.svg
	resetData string
	Reset     = unison.MustSVGFromContentString(resetData)

	//go:embed script.svg
	scriptData string
	Script     = unison.MustSVGFromContentString(scriptData)

	//go:embed settings.svg
	settingsData string
	Settings     = unison.MustSVGFromContentString(settingsData)

	//go:embed side_bar.svg
	sideBarData string
	SideBar     = unison.MustSVGFromContentString(sideBarData)

	//go:embed sign_post.svg
	signPostData string
	SignPost     = unison.MustSVGFromContentString(signPostData)

	//go:embed size_to_fit.svg
	sizeToFitData string
	SizeToFit     = unison.MustSVGFromContentString(sizeToFitData)

	//go:embed stack.svg
	stackData string
	Stack     = unison.MustSVGFromContentString(stackData)

	//go:embed stamper.svg
	stamperData string
	Stamper     = unison.MustSVGFromContentString(stamperData)

	//go:embed star.svg
	starData string
	Star     = unison.MustSVGFromContentString(starData)

	//go:embed trash.svg
	trashData string
	Trash     = unison.MustSVGFromContentString(trashData)

	//go:embed magic_wand.svg
	magicWandData string
	MagicWand     = unison.MustSVGFromContentString(magicWandData)

	//go:embed weight.svg
	weightData string
	Weight     = unison.MustSVGFromContentString(weightData)
)

// CreateImageFromSVG turns one of our svg-as-a-path objects into an actual SVG document, then renders it into an image
// at the specified square size. Note that this is not currently GPU accelerated, as I haven't added the necessary bits
// to unison to support scribbling into arbitrary offscreen images yet.
func CreateImageFromSVG(svg *unison.SVG, size int) (image.Image, error) {
	img, err := unison.NewImageFromDrawing(size, size, 72, func(gc *unison.Canvas) {
		svg.DrawInRectPreservingAspectRatio(gc, unison.NewRect(0, 0, float32(size), float32(size)), nil, nil)
	})
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return img.ToNRGBA()
}
