/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.template.Template;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.image.StdImageSet;
import com.trollworks.toolkit.utility.BundleInfo;
import com.trollworks.toolkit.utility.cmdline.CmdLine;
import com.trollworks.toolkit.utility.cmdline.CmdLineOption;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.util.jar.Attributes;

/** Provides standardized image access. */
@SuppressWarnings("nls")
public class GCSImages {
	static {
		StdImage.addLocation(GCSImages.class.getResource("images/"));
	}

	/** @return The exotic type icon. */
	public static final StdImage getExoticTypeIcon() {
		return StdImage.get("exotic_type");
	}

	/** @return The selected exotic type icon. */
	public static final StdImage getExoticTypeSelectedIcon() {
		return StdImage.get("exotic_type_selected");
	}

	/** @return The mental type icon. */
	public static final StdImage getMentalTypeIcon() {
		return StdImage.get("mental_type");
	}

	/** @return The selected mental type icon. */
	public static final StdImage getMentalTypeSelectedIcon() {
		return StdImage.get("mental_type_selected");
	}

	/** @return The physical type icon. */
	public static final StdImage getPhysicalTypeIcon() {
		return StdImage.get("physical_type");
	}

	/** @return The selected physical type icon. */
	public static final StdImage getPhysicalTypeSelectedIcon() {
		return StdImage.get("physical_type_selected");
	}

	/** @return The social type icon. */
	public static final StdImage getSocialTypeIcon() {
		return StdImage.get("social_type");
	}

	/** @return The selected social type icon. */
	public static final StdImage getSocialTypeSelectedIcon() {
		return StdImage.get("social_type_selected");
	}

	/** @return The supernatural type icon. */
	public static final StdImage getSupernaturalTypeIcon() {
		return StdImage.get("supernatural_type");
	}

	/** @return The selected supernatural type icon. */
	public static final StdImage getSupernaturalTypeSelectedIcon() {
		return StdImage.get("supernatural_type_selected");
	}

	/** @return The default portrait. */
	public static final StdImage getDefaultPortrait() {
		return StdImage.get("default_portrait");
	}

	/** @return The 'about' image. */
	public static final StdImage getAbout() {
		return StdImage.get("about");
	}

	/** @return The application icons. */
	public static final StdImageSet getAppIcons() {
		return StdImageSet.getOrLoad("app");
	}

	/** @return The character sheet icons. */
	public static final StdImageSet getCharacterSheetIcons() {
		return StdImageSet.getOrLoad(GURPSCharacter.EXTENSION);
	}

	/** @return The character template icons. */
	public static final StdImageSet getTemplateIcons() {
		return StdImageSet.getOrLoad(Template.EXTENSION);
	}

	/** @return The advantages icons. */
	public static final StdImageSet getAdvantagesIcons() {
		return StdImageSet.getOrLoad(AdvantageList.EXTENSION);
	}

	/** @return The skills icons. */
	public static final StdImageSet getSkillsIcons() {
		return StdImageSet.getOrLoad(SkillList.EXTENSION);
	}

	/** @return The spells icons. */
	public static final StdImageSet getSpellsIcons() {
		return StdImageSet.getOrLoad(SpellList.EXTENSION);
	}

	/** @return The equipment icons. */
	public static final StdImageSet getEquipmentIcons() {
		return StdImageSet.getOrLoad(EquipmentList.EXTENSION);
	}

	/** @return The character sheet icons. */
	public static final StdImageSet getCharacterSheetDocumentIcons() {
		return getDocumentIcons(GURPSCharacter.EXTENSION);
	}

	/** @return The character template icons. */
	public static final StdImageSet getTemplateDocumentIcons() {
		return getDocumentIcons(Template.EXTENSION);
	}

	/** @return The advantages icons. */
	public static final StdImageSet getAdvantagesDocumentIcons() {
		return getDocumentIcons(AdvantageList.EXTENSION);
	}

	/** @return The skills icons. */
	public static final StdImageSet getSkillsDocumentIcons() {
		return getDocumentIcons(SkillList.EXTENSION);
	}

	/** @return The spells icons. */
	public static final StdImageSet getSpellsDocumentIcons() {
		return getDocumentIcons(SpellList.EXTENSION);
	}

	/** @return The equipment icons. */
	public static final StdImageSet getEquipmentDocumentIcons() {
		return getDocumentIcons(EquipmentList.EXTENSION);
	}

	private static StdImageSet getDocumentIcons(String prefix) {
		String name = prefix + "_doc";
		StdImageSet set = StdImageSet.get(name);
		if (set == null) {
			set = new StdImageSet(name, StdImageSet.getOrLoad("document"), StdImageSet.getOrLoad(prefix));
		}
		return set;
	}

	/** Utility for creating GCS's icon sets. */
	public static void main(String[] args) {
		String name = "GenerateIcons";
		Attributes attributes = new Attributes();
		attributes.putValue(BundleInfo.BUNDLE_NAME, name);
		attributes.putValue(BundleInfo.BUNDLE_VERSION, "1.0");
		attributes.putValue(BundleInfo.BUNDLE_COPYRIGHT_OWNER, "Richard A. Wilkes");
		attributes.putValue(BundleInfo.BUNDLE_COPYRIGHT_YEARS, "2014");
		attributes.putValue(BundleInfo.BUNDLE_LICENSE, "Mozilla Public License 2.0");
		BundleInfo.setDefault(new BundleInfo(attributes, name));
		CmdLineOption icnsOption = new CmdLineOption("Generate ICNS files", null, "icns");
		CmdLineOption icoOption = new CmdLineOption("Generate ICO files", null, "ico");
		CmdLineOption appOption = new CmdLineOption("Generate just the 128x128 app icon", null, "app");
		CmdLineOption dirOption = new CmdLineOption("The directory to place the generated files into", "DIR", "dir");
		CmdLine cmdline = new CmdLine();
		cmdline.addOptions(icnsOption, icoOption, appOption, dirOption);
		cmdline.processArguments(args);
		boolean icns = cmdline.isOptionUsed(icnsOption);
		boolean ico = cmdline.isOptionUsed(icoOption);
		boolean app = cmdline.isOptionUsed(appOption);
		if (!icns && !ico && !app) {
			System.err.printf("At least one of %s, %s, or %s must be specified.\n", icnsOption, icoOption, appOption);
			System.exit(1);
		}
		try {
			File dir = new File(cmdline.isOptionUsed(dirOption) ? cmdline.getOptionArgument(dirOption) : ".");
			System.out.println("Generating icons into " + dir);
			dir.mkdirs();
			if (app) {
				File file = new File(dir, "gcs.png");
				if (StdImage.writePNG(file, getAppIcons().getImage(128), 72)) {
					System.out.println("Created: " + file);
				} else {
					System.err.println("Unable to create: " + file);
				}
			}
			if (icns || ico) {
				createIconFiles(getAppIcons(), dir, "app", icns, ico);
				createIconFiles(getAdvantagesDocumentIcons(), dir, AdvantageList.EXTENSION, icns, ico);
				createIconFiles(getEquipmentDocumentIcons(), dir, EquipmentList.EXTENSION, icns, ico);
				createIconFiles(getCharacterSheetDocumentIcons(), dir, GURPSCharacter.EXTENSION, icns, ico);
				createIconFiles(getTemplateDocumentIcons(), dir, Template.EXTENSION, icns, ico);
				createIconFiles(getSkillsDocumentIcons(), dir, SkillList.EXTENSION, icns, ico);
				createIconFiles(getSpellsDocumentIcons(), dir, SpellList.EXTENSION, icns, ico);
			}
		} catch (Exception exception) {
			exception.printStackTrace(System.err);
		}
	}

	private static void createIconFiles(StdImageSet set, File dir, String name, boolean generateICNS, boolean generateICO) throws IOException {
		for (int size : StdImageSet.STD_SIZES) {
			set.getImage(size);
		}
		File file;
		if (generateICNS) {
			file = new File(dir, name + ".icns");
			try (BufferedOutputStream out = new BufferedOutputStream(new FileOutputStream(file))) {
				set.saveAsIcns(out);
				System.out.println("Created: " + file);
			}
		}
		if (generateICO) {
			file = new File(dir, name + ".ico");
			try (BufferedOutputStream out = new BufferedOutputStream(new FileOutputStream(file))) {
				set.saveAsIco(out);
				System.out.println("Created: " + file);
			}
		}
	}
}
