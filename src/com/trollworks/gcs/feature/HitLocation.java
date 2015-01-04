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

package com.trollworks.gcs.feature;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

import java.util.ArrayList;

/** Hit locations. */
public enum HitLocation {
	/** The skull hit location. */
	SKULL {
		@Override
		public String toString() {
			return SKULL_TITLE;
		}
	},
	/** The eyes hit location. */
	EYES {
		@Override
		public String toString() {
			return EYES_TITLE;
		}
	},
	/** The face hit location. */
	FACE {
		@Override
		public String toString() {
			return FACE_TITLE;
		}
	},
	/** The neck hit location. */
	NECK {
		@Override
		public String toString() {
			return NECK_TITLE;
		}
	},
	/** The torso hit location. */
	TORSO {
		@Override
		public String toString() {
			return TORSO_TITLE;
		}
	},
	/** The vitals hit location. */
	VITALS {
		@Override
		public String toString() {
			return VITALS_TITLE;
		}

		@Override
		public boolean isChoosable() {
			return false;
		}
	},
	/** The groin hit location. */
	GROIN {
		@Override
		public String toString() {
			return GROIN_TITLE;
		}
	},
	/** The arm hit location. */
	ARMS {
		@Override
		public String toString() {
			return ARMS_TITLE;
		}
	},
	/** The hand hit location. */
	HANDS {
		@Override
		public String toString() {
			return HANDS_TITLE;
		}
	},
	/** The leg hit location. */
	LEGS {
		@Override
		public String toString() {
			return LEGS_TITLE;
		}
	},
	/** The foot hit location. */
	FEET {
		@Override
		public String toString() {
			return FEET_TITLE;
		}
	},
	/** The full body hit location. */
	FULL_BODY {
		@Override
		public String toString() {
			return FULL_BODY_TITLE;
		}
	},
	/** The full body except eyes hit location. */
	FULL_BODY_EXCEPT_EYES {
		@Override
		public String toString() {
			return FULL_BODY_EXCEPT_EYES_TITLE;
		}
	};

	@Localize("to the skull")
	@Localize(locale = "de", value = "auf den Schädel")
	@Localize(locale = "ru", value = "черепу")
	static String	SKULL_TITLE;
	@Localize("to the eyes")
	@Localize(locale = "de", value = "auf die Augen")
	@Localize(locale = "ru", value = "глазам")
	static String	EYES_TITLE;
	@Localize("to the face")
	@Localize(locale = "de", value = "auf das Gesicht")
	@Localize(locale = "ru", value = "лицу")
	static String	FACE_TITLE;
	@Localize("to the neck")
	@Localize(locale = "de", value = "auf den Hals")
	@Localize(locale = "ru", value = "шее")
	static String	NECK_TITLE;
	@Localize("to the torso")
	@Localize(locale = "de", value = "auf den Torso")
	@Localize(locale = "ru", value = "туловищу")
	static String	TORSO_TITLE;
	@Localize("to the vitals")
	@Localize(locale = "de", value = "auf die Organe")
	@Localize(locale = "ru", value = "жизненно-важным органам")
	static String	VITALS_TITLE;
	@Localize("to the groin")
	@Localize(locale = "de", value = "auf die Leiste")
	@Localize(locale = "ru", value = "паху")
	static String	GROIN_TITLE;
	@Localize("to the arms")
	@Localize(locale = "de", value = "auf die Arme")
	@Localize(locale = "ru", value = "рукам")
	static String	ARMS_TITLE;
	@Localize("to the hands")
	@Localize(locale = "de", value = "auf die Hände")
	@Localize(locale = "ru", value = "рукам")
	static String	HANDS_TITLE;
	@Localize("to the legs")
	@Localize(locale = "de", value = "auf die Beine")
	@Localize(locale = "ru", value = "ногам")
	static String	LEGS_TITLE;
	@Localize("to the feet")
	@Localize(locale = "de", value = "auf die Füße")
	@Localize(locale = "ru", value = "стопам")
	static String	FEET_TITLE;
	@Localize("to the full body")
	@Localize(locale = "de", value = "auf den gesamten Körper")
	@Localize(locale = "ru", value = "всему телу")
	static String	FULL_BODY_TITLE;
	@Localize("to the full body except the eyes")
	@Localize(locale = "de", value = "auf den gesamten Körper ohne Augen")
	@Localize(locale = "ru", value = "всему телу, кроме глаз")
	static String	FULL_BODY_EXCEPT_EYES_TITLE;

	static {
		Localization.initialize();
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
