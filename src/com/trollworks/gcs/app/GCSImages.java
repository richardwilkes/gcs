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

import com.trollworks.toolkit.ui.image.Images;

import java.awt.image.BufferedImage;

/** Provides standardized image access. */
public class GCSImages {
	private static final String	SMALL	= "Small";	//$NON-NLS-1$
	private static final String	LARGE	= "Large";	//$NON-NLS-1$
	private static final String	SINGLE	= "Single"; //$NON-NLS-1$
	private static final String	FILE	= "";		//$NON-NLS-1$

	static {
		Images.addLocation(GCSImages.class.getResource("images/")); //$NON-NLS-1$
	}

	/** @return The exotic type icon. */
	public static final BufferedImage getExoticTypeIcon() {
		return Images.get("ExoticType"); //$NON-NLS-1$
	}

	/** @return The selected exotic type icon. */
	public static final BufferedImage getExoticTypeSelectedIcon() {
		return Images.get("ExoticTypeSelected"); //$NON-NLS-1$
	}

	/** @return The mental type icon. */
	public static final BufferedImage getMentalTypeIcon() {
		return Images.get("MentalType"); //$NON-NLS-1$
	}

	/** @return The selected mental type icon. */
	public static final BufferedImage getMentalTypeSelectedIcon() {
		return Images.get("MentalTypeSelected"); //$NON-NLS-1$
	}

	/** @return The physical type icon. */
	public static final BufferedImage getPhysicalTypeIcon() {
		return Images.get("PhysicalType"); //$NON-NLS-1$
	}

	/** @return The selected physical type icon. */
	public static final BufferedImage getPhysicalTypeSelectedIcon() {
		return Images.get("PhysicalTypeSelected"); //$NON-NLS-1$
	}

	/** @return The social type icon. */
	public static final BufferedImage getSocialTypeIcon() {
		return Images.get("SocialType"); //$NON-NLS-1$
	}

	/** @return The selected social type icon. */
	public static final BufferedImage getSocialTypeSelectedIcon() {
		return Images.get("SocialTypeSelected"); //$NON-NLS-1$
	}

	/** @return The supernatural type icon. */
	public static final BufferedImage getSupernaturalTypeIcon() {
		return Images.get("SupernaturalType"); //$NON-NLS-1$
	}

	/** @return The selected supernatural type icon. */
	public static final BufferedImage getSupernaturalTypeSelectedIcon() {
		return Images.get("SupernaturalTypeSelected"); //$NON-NLS-1$
	}

	/** @return The default portrait. */
	public static final BufferedImage getDefaultPortrait() {
		return Images.get("DefaultPortrait"); //$NON-NLS-1$
	}

	/** @return The default window icon. */
	public static final BufferedImage getDefaultWindowIcon() {
		return Images.get("DefaultWindowIcon"); //$NON-NLS-1$
	}

	/** @return The splash image. */
	public static final BufferedImage getSplash() {
		return Images.get("Splash"); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @return The character sheet icon.
	 */
	public static final BufferedImage getCharacterSheetIcon(boolean large) {
		return Images.get("CharacterSheet" + (large ? LARGE : SMALL)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @return The template icon.
	 */
	public static final BufferedImage getTemplateIcon(boolean large) {
		return Images.get("Template" + (large ? LARGE : SMALL)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The advantage icon.
	 */
	public static final BufferedImage getAdvantageIcon(boolean large, boolean single) {
		return Images.get("Advantage" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The skill icon.
	 */
	public static final BufferedImage getSkillIcon(boolean large, boolean single) {
		return Images.get("Skill" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The spell icon.
	 */
	public static final BufferedImage getSpellIcon(boolean large, boolean single) {
		return Images.get("Spell" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The equipment icon.
	 */
	public static final BufferedImage getEquipmentIcon(boolean large, boolean single) {
		return Images.get("Equipment" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @return The library icon.
	 */
	public static final BufferedImage getLibraryIcon(boolean large) {
		return Images.get("Library" + (large ? LARGE : SMALL)); //$NON-NLS-1$
	}
}
