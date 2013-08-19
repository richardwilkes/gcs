/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.app;

import com.trollworks.ttk.image.Images;

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

	/**
	 * @param name The name to search for.
	 * @return The image for the specified name.
	 */
	public static final synchronized BufferedImage get(String name) {
		return Images.get(name, true);
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
