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

/** Hit locations. */
public enum CMHitLocation {
	/** The skull hit location. */
	SKULL(Msgs.SKULL),
	/** The eyes hit location. */
	EYES(Msgs.EYES),
	/** The face hit location. */
	FACE(Msgs.FACE),
	/** The neck hit location. */
	NECK(Msgs.NECK),
	/** The torso hit location. */
	TORSO(Msgs.TORSO),
	/** The vitals hit location. */
	VITALS(Msgs.VITALS) {
		@Override public boolean isChoosable() {
			return false;
		}
	},
	/** The groin hit location. */
	GROIN(Msgs.GROIN),
	/** The arm hit location. */
	ARMS(Msgs.ARMS),
	/** The hand hit location. */
	HANDS(Msgs.HANDS),
	/** The leg hit location. */
	LEGS(Msgs.LEGS),
	/** The foot hit location. */
	FEET(Msgs.FEET),
	/** The full body hit location. */
	FULL_BODY(Msgs.FULL_BODY),
	/** The full body except eyes hit location. */
	FULL_BODY_EXCEPT_EYES(Msgs.FULL_BODY_EXCEPT_EYES);

	private String mTitle;

	private CMHitLocation(String title) {
		mTitle = title;
	}

	@Override public String toString() {
		return mTitle;
	}

	/** @return Whether this location is choosable as an armor protection spot. */
	public boolean isChoosable() {
		return true;
	}
}
