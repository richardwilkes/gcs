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

import com.trollworks.toolkit.ui.image.IconSet;
import com.trollworks.toolkit.ui.image.Images;
import com.trollworks.toolkit.ui.image.ToolkitIcon;

/** Provides standardized image access. */
@SuppressWarnings("nls")
public class GCSImages {
	private static final String	SMALL	= "Small";
	private static final String	LARGE	= "Large";
	private static final String	SINGLE	= "Single";
	private static final String	FILE	= "";

	static {
		Images.addLocation(GCSImages.class.getResource("images/"));
	}

	/** @return The exotic type icon. */
	public static final ToolkitIcon getExoticTypeIcon() {
		return Images.get("ExoticType");
	}

	/** @return The selected exotic type icon. */
	public static final ToolkitIcon getExoticTypeSelectedIcon() {
		return Images.get("ExoticTypeSelected");
	}

	/** @return The mental type icon. */
	public static final ToolkitIcon getMentalTypeIcon() {
		return Images.get("MentalType");
	}

	/** @return The selected mental type icon. */
	public static final ToolkitIcon getMentalTypeSelectedIcon() {
		return Images.get("MentalTypeSelected");
	}

	/** @return The physical type icon. */
	public static final ToolkitIcon getPhysicalTypeIcon() {
		return Images.get("PhysicalType");
	}

	/** @return The selected physical type icon. */
	public static final ToolkitIcon getPhysicalTypeSelectedIcon() {
		return Images.get("PhysicalTypeSelected");
	}

	/** @return The social type icon. */
	public static final ToolkitIcon getSocialTypeIcon() {
		return Images.get("SocialType");
	}

	/** @return The selected social type icon. */
	public static final ToolkitIcon getSocialTypeSelectedIcon() {
		return Images.get("SocialTypeSelected");
	}

	/** @return The supernatural type icon. */
	public static final ToolkitIcon getSupernaturalTypeIcon() {
		return Images.get("SupernaturalType");
	}

	/** @return The selected supernatural type icon. */
	public static final ToolkitIcon getSupernaturalTypeSelectedIcon() {
		return Images.get("SupernaturalTypeSelected");
	}

	/** @return The default portrait. */
	public static final ToolkitIcon getDefaultPortrait() {
		return Images.get("DefaultPortrait");
	}

	/** @return The 'about' image. */
	public static final ToolkitIcon getAbout() {
		return Images.get("About");
	}

	/** @return The application icons. */
	public static final IconSet getAppIcons() {
		return IconSet.get("AppIcon");
	}

	/** @return The character sheet icons. */
	public static final IconSet getCharacterSheetIcons() {
		return IconSet.get("CharacterSheet");
	}

	/** @return The character template icons. */
	public static final IconSet getTemplateIcons() {
		return IconSet.get("Template");
	}

	/** @return The library icons. */
	public static final IconSet getLibraryIcons() {
		return IconSet.get("Library");
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The advantage icon.
	 */
	public static final ToolkitIcon getAdvantageIcon(boolean large, boolean single) {
		return Images.get("Advantage" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The skill icon.
	 */
	public static final ToolkitIcon getSkillIcon(boolean large, boolean single) {
		return Images.get("Skill" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The spell icon.
	 */
	public static final ToolkitIcon getSpellIcon(boolean large, boolean single) {
		return Images.get("Spell" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The equipment icon.
	 */
	public static final ToolkitIcon getEquipmentIcon(boolean large, boolean single) {
		return Images.get("Equipment" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}
}
