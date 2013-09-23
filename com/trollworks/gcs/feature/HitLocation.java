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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;

import java.util.ArrayList;

@Localized({
				@LS(key = "SKULL", msg = "to the skull"),
				@LS(key = "EYES", msg = "to the eyes"),
				@LS(key = "FACE", msg = "to the face"),
				@LS(key = "NECK", msg = "to the neck"),
				@LS(key = "TORSO", msg = "to the torso"),
				@LS(key = "VITALS", msg = "to the vitals"),
				@LS(key = "GROIN", msg = "to the groin"),
				@LS(key = "ARMS", msg = "to the arms"),
				@LS(key = "HANDS", msg = "to the hands"),
				@LS(key = "LEGS", msg = "to the legs"),
				@LS(key = "FEET", msg = "to the feet"),
				@LS(key = "FULL_BODY", msg = "to the full body"),
				@LS(key = "FULL_BODY_EXCEPT_EYES", msg = "to the full body except the eyes"),
})
/** Hit locations. */
public enum HitLocation {
	/** The skull hit location. */
	SKULL,
	/** The eyes hit location. */
	EYES,
	/** The face hit location. */
	FACE,
	/** The neck hit location. */
	NECK,
	/** The torso hit location. */
	TORSO,
	/** The vitals hit location. */
	VITALS {
		@Override
		public boolean isChoosable() {
			return false;
		}
	},
	/** The groin hit location. */
	GROIN,
	/** The arm hit location. */
	ARMS,
	/** The hand hit location. */
	HANDS,
	/** The leg hit location. */
	LEGS,
	/** The foot hit location. */
	FEET,
	/** The full body hit location. */
	FULL_BODY,
	/** The full body except eyes hit location. */
	FULL_BODY_EXCEPT_EYES;

	@Override
	public String toString() {
		return HitLocation_LS.toString(this);
	}

	/** @return The hit locations that can be chosen as an armor protection spot. */
	public static HitLocation[] getChoosableLocations() {
		ArrayList<HitLocation> list = new ArrayList<>();
		for (HitLocation one : values()) {
			if (one.isChoosable()) {
				list.add(one);
			}
		}
		return list.toArray(new HitLocation[list.size()]);
	}

	/** @return Whether this location is choosable as an armor protection spot. */
	@SuppressWarnings("static-method")
	public boolean isChoosable() {
		return true;
	}
}
