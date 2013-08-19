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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.ttk.utility.LocalizedMessages;

import java.util.ArrayList;

/** Hit locations. */
public enum HitLocation {
	/** The skull hit location. */
	SKULL {
		@Override public String toString() {
			return MSG_SKULL;
		}
	},
	/** The eyes hit location. */
	EYES {
		@Override public String toString() {
			return MSG_EYES;
		}
	},
	/** The face hit location. */
	FACE {
		@Override public String toString() {
			return MSG_FACE;
		}
	},
	/** The neck hit location. */
	NECK {
		@Override public String toString() {
			return MSG_NECK;
		}
	},
	/** The torso hit location. */
	TORSO {
		@Override public String toString() {
			return MSG_TORSO;
		}
	},
	/** The vitals hit location. */
	VITALS {
		@Override public String toString() {
			return MSG_VITALS;
		}

		@Override public boolean isChoosable() {
			return false;
		}
	},
	/** The groin hit location. */
	GROIN {
		@Override public String toString() {
			return MSG_GROIN;
		}
	},
	/** The arm hit location. */
	ARMS {
		@Override public String toString() {
			return MSG_ARMS;
		}
	},
	/** The hand hit location. */
	HANDS {
		@Override public String toString() {
			return MSG_HANDS;
		}
	},
	/** The leg hit location. */
	LEGS {
		@Override public String toString() {
			return MSG_LEGS;
		}
	},
	/** The foot hit location. */
	FEET {
		@Override public String toString() {
			return MSG_FEET;
		}
	},
	/** The full body hit location. */
	FULL_BODY {
		@Override public String toString() {
			return MSG_FULL_BODY;
		}
	},
	/** The full body except eyes hit location. */
	FULL_BODY_EXCEPT_EYES {
		@Override public String toString() {
			return MSG_FULL_BODY_EXCEPT_EYES;
		}
	};

	static String	MSG_SKULL;
	static String	MSG_EYES;
	static String	MSG_FACE;
	static String	MSG_NECK;
	static String	MSG_TORSO;
	static String	MSG_FULL_BODY;
	static String	MSG_FULL_BODY_EXCEPT_EYES;
	static String	MSG_VITALS;
	static String	MSG_GROIN;
	static String	MSG_ARMS;
	static String	MSG_HANDS;
	static String	MSG_LEGS;
	static String	MSG_FEET;

	static {
		LocalizedMessages.initialize(HitLocation.class);
	}

	/** @return The hit locations that can be chosen as an armor protection spot. */
	public static HitLocation[] getChoosableLocations() {
		ArrayList<HitLocation> list = new ArrayList<HitLocation>();
		for (HitLocation one : values()) {
			if (one.isChoosable()) {
				list.add(one);
			}
		}
		return list.toArray(new HitLocation[list.size()]);
	}

	/** @return Whether this location is choosable as an armor protection spot. */
	public boolean isChoosable() {
		return true;
	}
}
