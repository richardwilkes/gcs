/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
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
	/** The tail hit location. */
	TAIL {
		@Override
		public String toString() {
			return TAIL_TITLE;
		}
	},
	/** The wing hit location. */
	WINGS {
		@Override
		public String toString() {
			return WINGS_TITLE;
		}
	},
	/** The fin hit location. */
	FINS {
		@Override
		public String toString() {
			return FINS_TITLE;
		}
	},
	/** The brain hit location. */
	BRAIN {
		@Override
		public String toString() {
			return BRAIN_TITLE;
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
	@Localize(locale = "es", value = "en el cráneo")
	static String	SKULL_TITLE;
	@Localize("to the eyes")
	@Localize(locale = "de", value = "auf die Augen")
	@Localize(locale = "ru", value = "глазам")
	@Localize(locale = "es", value = "en los ojos")
	static String	EYES_TITLE;
	@Localize("to the face")
	@Localize(locale = "de", value = "auf das Gesicht")
	@Localize(locale = "ru", value = "лицу")
	@Localize(locale = "es", value = "en la cara")
	static String	FACE_TITLE;
	@Localize("to the neck")
	@Localize(locale = "de", value = "auf den Hals")
	@Localize(locale = "ru", value = "шее")
	@Localize(locale = "es", value = "en el cuello")
	static String	NECK_TITLE;
	@Localize("to the torso")
	@Localize(locale = "de", value = "auf den Torso")
	@Localize(locale = "ru", value = "туловищу")
	@Localize(locale = "es", value = "en el torso")
	static String	TORSO_TITLE;
	@Localize("to the vitals")
	@Localize(locale = "de", value = "auf die Organe")
	@Localize(locale = "ru", value = "жизненно-важным органам")
	@Localize(locale = "es", value = "en los órganos vitales")
	static String	VITALS_TITLE;
	@Localize("to the groin")
	@Localize(locale = "de", value = "auf die Leiste")
	@Localize(locale = "ru", value = "паху")
	@Localize(locale = "es", value = "en las ingles")
	static String	GROIN_TITLE;
	@Localize("to the arms")
	@Localize(locale = "de", value = "auf die Arme")
	@Localize(locale = "ru", value = "рукам")
	@Localize(locale = "es", value = "en los brazos")
	static String	ARMS_TITLE;
	@Localize("to the hands")
	@Localize(locale = "de", value = "auf die Hände")
	@Localize(locale = "ru", value = "рукам")
	@Localize(locale = "es", value = "en las manos")
	static String	HANDS_TITLE;
	@Localize("to the legs")
	@Localize(locale = "de", value = "auf die Beine")
	@Localize(locale = "ru", value = "ногам")
	@Localize(locale = "es", value = "en las piernas")
	static String	LEGS_TITLE;
	@Localize("to the feet")
	@Localize(locale = "de", value = "auf die Füße")
	@Localize(locale = "ru", value = "стопам")
	@Localize(locale = "es", value = "en los pies")
	static String	FEET_TITLE;
	@Localize("to the tail")
	static String	TAIL_TITLE;
	@Localize("to the wings")
	static String	WINGS_TITLE;
	@Localize("to the fins")
	static String	FINS_TITLE;
	@Localize("to the brain")
	static String	BRAIN_TITLE;
	@Localize("to the full body")
	@Localize(locale = "de", value = "auf den gesamten Körper")
	@Localize(locale = "ru", value = "всему телу")
	@Localize(locale = "es", value = "en todo el cuerpo")
	static String	FULL_BODY_TITLE;
	@Localize("to the full body except the eyes")
	@Localize(locale = "de", value = "auf den gesamten Körper ohne Augen")
	@Localize(locale = "ru", value = "всему телу, кроме глаз")
	@Localize(locale = "es", value = "en todo el cuerpo salvo en los ojos")
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
