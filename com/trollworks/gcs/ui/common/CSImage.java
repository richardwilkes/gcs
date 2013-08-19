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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.common;

import com.trollworks.toolkit.io.TKImage;

import java.awt.image.BufferedImage;

/** Image accessors. */
public class CSImage {
	private static final String	SMALL	= "Small";	//$NON-NLS-1$
	private static final String	LARGE	= "Large";	//$NON-NLS-1$
	private static final String	SINGLE	= "Single"; //$NON-NLS-1$
	private static final String	FILE	= "";		//$NON-NLS-1$

	/** @return The more icon. */
	public static final BufferedImage getMoreIcon() {
		return TKImage.get("CSMore"); //$NON-NLS-1$
	}

	/** @return The add icon. */
	public static final BufferedImage getAddIcon() {
		return TKImage.get("CSAdd"); //$NON-NLS-1$
	}

	/** @return The remove icon. */
	public static final BufferedImage getRemoveIcon() {
		return TKImage.get("CSRemove"); //$NON-NLS-1$
	}

	/** @return The locked icon. */
	public static final BufferedImage getLockedIcon() {
		return TKImage.get("CSLocked"); //$NON-NLS-1$
	}

	/** @return The unlocked icon. */
	public static final BufferedImage getUnlockedIcon() {
		return TKImage.get("CSUnlocked"); //$NON-NLS-1$
	}

	/** @return The exotic type icon. */
	public static final BufferedImage getExoticTypeIcon() {
		return TKImage.get("CSExoticType"); //$NON-NLS-1$
	}

	/** @return The selected exotic type icon. */
	public static final BufferedImage getExoticTypeSelectedIcon() {
		return TKImage.get("CSExoticTypeSelected"); //$NON-NLS-1$
	}

	/** @return The mental type icon. */
	public static final BufferedImage getMentalTypeIcon() {
		return TKImage.get("CSMentalType"); //$NON-NLS-1$
	}

	/** @return The selected mental type icon. */
	public static final BufferedImage getMentalTypeSelectedIcon() {
		return TKImage.get("CSMentalTypeSelected"); //$NON-NLS-1$
	}

	/** @return The physical type icon. */
	public static final BufferedImage getPhysicalTypeIcon() {
		return TKImage.get("CSPhysicalType"); //$NON-NLS-1$
	}

	/** @return The selected physical type icon. */
	public static final BufferedImage getPhysicalTypeSelectedIcon() {
		return TKImage.get("CSPhysicalTypeSelected"); //$NON-NLS-1$
	}

	/** @return The social type icon. */
	public static final BufferedImage getSocialTypeIcon() {
		return TKImage.get("CSSocialType"); //$NON-NLS-1$
	}

	/** @return The selected social type icon. */
	public static final BufferedImage getSocialTypeSelectedIcon() {
		return TKImage.get("CSSocialTypeSelected"); //$NON-NLS-1$
	}

	/** @return The supernatural type icon. */
	public static final BufferedImage getSupernaturalTypeIcon() {
		return TKImage.get("CSSupernaturalType"); //$NON-NLS-1$
	}

	/** @return The selected supernatural type icon. */
	public static final BufferedImage getSupernaturalTypeSelectedIcon() {
		return TKImage.get("CSSupernaturalTypeSelected"); //$NON-NLS-1$
	}

	/** @return The default portrait. */
	public static final BufferedImage getDefaultPortrait() {
		return TKImage.get("CSDefaultPortrait"); //$NON-NLS-1$
	}

	/** @return The default window icon. */
	public static final BufferedImage getDefaultWindowIcon() {
		return TKImage.get("CSDefaultWindowIcon"); //$NON-NLS-1$
	}

	/** @return The splash image. */
	public static final BufferedImage getSplash() {
		return TKImage.get("CSSplash"); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @return The character sheet icon.
	 */
	public static final BufferedImage getCharacterSheetIcon(boolean large) {
		return TKImage.get(large ? "CSCharacterSheetLarge" : "CSCharacterSheetSmall"); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @return The template icon.
	 */
	public static final BufferedImage getTemplateIcon(boolean large) {
		return TKImage.get(large ? "CSTemplateLarge" : "CSTemplateSmall"); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The advantage icon.
	 */
	public static final BufferedImage getAdvantageIcon(boolean large, boolean single) {
		return TKImage.get("CSAdvantage" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The skill icon.
	 */
	public static final BufferedImage getSkillIcon(boolean large, boolean single) {
		return TKImage.get("CSSkill" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The spell icon.
	 */
	public static final BufferedImage getSpellIcon(boolean large, boolean single) {
		return TKImage.get("CSSpell" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The equipment icon.
	 */
	public static final BufferedImage getEquipmentIcon(boolean large, boolean single) {
		return TKImage.get("CSEquipment" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE)); //$NON-NLS-1$
	}
}
