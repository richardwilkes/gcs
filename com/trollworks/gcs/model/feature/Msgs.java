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

package com.trollworks.gcs.model.feature;

import com.trollworks.toolkit.io.TKMessages;

/** All localized strings in this package should be accessed via this class. */
class Msgs {
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	ST;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	DX;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	IQ;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	HT;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	WILL;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	PERCEPTION;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	VISION;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	HEARING;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	TASTE_SMELL;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	TOUCH;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	DODGE;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	PARRY;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	BLOCK;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	SPEED;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	MOVE;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	FP;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	HP;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	SM;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	NONE;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	STRIKING_ONLY;
	/** Used by {@link CMAttributeBonusLimitation}. */
	public static String	LIFTING_ONLY;

	/** Used by {@link CMHitLocation}. */
	public static String	SKULL;
	/** Used by {@link CMHitLocation}. */
	public static String	EYES;
	/** Used by {@link CMHitLocation}. */
	public static String	FACE;
	/** Used by {@link CMHitLocation}. */
	public static String	NECK;
	/** Used by {@link CMHitLocation}. */
	public static String	TORSO;
	/** Used by {@link CMHitLocation}. */
	public static String	FULL_BODY;
	/** Used by {@link CMHitLocation}. */
	public static String	FULL_BODY_EXCEPT_EYES;
	/** Used by {@link CMHitLocation}. */
	public static String	VITALS;
	/** Used by {@link CMHitLocation}. */
	public static String	GROIN;
	/** Used by {@link CMHitLocation}. */
	public static String	ARMS;
	/** Used by {@link CMHitLocation}. */
	public static String	HANDS;
	/** Used by {@link CMHitLocation}. */
	public static String	LEGS;
	/** Used by {@link CMHitLocation}. */
	public static String	FEET;

	static {
		TKMessages.initialize(Msgs.class);
	}
}
