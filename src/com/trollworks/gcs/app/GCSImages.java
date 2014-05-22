/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.toolkit.ui.image.IconSet;
import com.trollworks.toolkit.ui.image.Images;
import com.trollworks.toolkit.ui.image.ToolkitIcon;
import com.trollworks.toolkit.utility.BundleInfo;
import com.trollworks.toolkit.utility.cmdline.CmdLine;
import com.trollworks.toolkit.utility.cmdline.CmdLineOption;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;

/** Provides standardized image access. */
@SuppressWarnings("nls")
public class GCSImages {
	static {
		Images.addLocation(GCSImages.class.getResource("images/"));
	}

	/** @return The exotic type icon. */
	public static final ToolkitIcon getExoticTypeIcon() {
		return Images.get("exotic_type");
	}

	/** @return The selected exotic type icon. */
	public static final ToolkitIcon getExoticTypeSelectedIcon() {
		return Images.get("exotic_type_selected");
	}

	/** @return The mental type icon. */
	public static final ToolkitIcon getMentalTypeIcon() {
		return Images.get("mental_type");
	}

	/** @return The selected mental type icon. */
	public static final ToolkitIcon getMentalTypeSelectedIcon() {
		return Images.get("mental_type_selected");
	}

	/** @return The physical type icon. */
	public static final ToolkitIcon getPhysicalTypeIcon() {
		return Images.get("physical_type");
	}

	/** @return The selected physical type icon. */
	public static final ToolkitIcon getPhysicalTypeSelectedIcon() {
		return Images.get("physical_type_selected");
	}

	/** @return The social type icon. */
	public static final ToolkitIcon getSocialTypeIcon() {
		return Images.get("social_type");
	}

	/** @return The selected social type icon. */
	public static final ToolkitIcon getSocialTypeSelectedIcon() {
		return Images.get("social_type_selected");
	}

	/** @return The supernatural type icon. */
	public static final ToolkitIcon getSupernaturalTypeIcon() {
		return Images.get("supernatural_type");
	}

	/** @return The selected supernatural type icon. */
	public static final ToolkitIcon getSupernaturalTypeSelectedIcon() {
		return Images.get("supernatural_type_selected");
	}

	/** @return The default portrait. */
	public static final ToolkitIcon getDefaultPortrait() {
		return Images.get("default_portrait");
	}

	/** @return The 'about' image. */
	public static final ToolkitIcon getAbout() {
		return Images.get("about");
	}

	/** @return The application icons. */
	public static final IconSet getAppIcons() {
		return IconSet.getOrLoad("app");
	}

	/** @return The character sheet icons. */
	public static final IconSet getCharacterSheetIcons() {
		return IconSet.getOrLoad(GURPSCharacter.EXTENSION);
	}

	/** @return The character template icons. */
	public static final IconSet getTemplateIcons() {
		return IconSet.getOrLoad(Template.EXTENSION);
	}

	/** @return The advantages icons. */
	public static final IconSet getAdvantagesIcons() {
		return IconSet.getOrLoad(AdvantageList.EXTENSION);
	}

	/** @return The skills icons. */
	public static final IconSet getSkillsIcons() {
		return IconSet.getOrLoad(SkillList.EXTENSION);
	}

	/** @return The spells icons. */
	public static final IconSet getSpellsIcons() {
		return IconSet.getOrLoad(SpellList.EXTENSION);
	}

	/** @return The equipment icons. */
	public static final IconSet getEquipmentIcons() {
		return IconSet.getOrLoad(EquipmentList.EXTENSION);
	}

	/** @return The character sheet icons. */
	public static final IconSet getCharacterSheetDocumentIcons() {
		return getDocumentIcons(GURPSCharacter.EXTENSION);
	}

	/** @return The character template icons. */
	public static final IconSet getTemplateDocumentIcons() {
		return getDocumentIcons(Template.EXTENSION);
	}

	/** @return The advantages icons. */
	public static final IconSet getAdvantagesDocumentIcons() {
		return getDocumentIcons(AdvantageList.EXTENSION);
	}

	/** @return The skills icons. */
	public static final IconSet getSkillsDocumentIcons() {
		return getDocumentIcons(SkillList.EXTENSION);
	}

	/** @return The spells icons. */
	public static final IconSet getSpellsDocumentIcons() {
		return getDocumentIcons(SpellList.EXTENSION);
	}

	/** @return The equipment icons. */
	public static final IconSet getEquipmentDocumentIcons() {
		return getDocumentIcons(EquipmentList.EXTENSION);
	}

	private static IconSet getDocumentIcons(String prefix) {
		String name = prefix + "_doc";
		IconSet set = IconSet.get(name);
		if (set == null) {
			set = new IconSet(name, IconSet.getOrLoad("document"), IconSet.getOrLoad(prefix));
		}
		return set;
	}

	/** Utility for creating GCS's icon sets. */
	public static void main(String[] args) {
		BundleInfo.setDefault(new BundleInfo("GenerateIcons", "1.0", "Richard A. Wilkes", "2014", "Mozilla Public License 2.0"));
		CmdLineOption icnsOption = new CmdLineOption("Generate ICNS files", null, "icns");
		CmdLineOption icoOption = new CmdLineOption("Generate ICO files", null, "ico");
		CmdLineOption appOption = new CmdLineOption("Generate just the 128x128 app icon", null, "app");
		CmdLineOption dirOption = new CmdLineOption("The directory to place the generated files into", "DIR", "dir");
		CmdLine cmdline = new CmdLine();
		cmdline.addOptions(icnsOption, icoOption, dirOption);
		cmdline.processArguments(args);
		boolean icns = cmdline.isOptionUsed(icnsOption);
		boolean ico = cmdline.isOptionUsed(icoOption);
		boolean app = cmdline.isOptionUsed(icoOption);
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
				if (Images.writePNG(file, getAppIcons().getIcon(128), 72)) {
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

	private static void createIconFiles(IconSet set, File dir, String name, boolean generateICNS, boolean generateICO) throws IOException {
		for (int size : IconSet.STD_SIZES) {
			set.getIcon(size);
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
