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

package com.trollworks.gcs.skill;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** The possible skill difficulty levels. */
public enum SkillDifficulty {
	/** The "easy" difficulty. */
	E {
		@Override
		public String toString() {
			return E_TITLE;
		}
	},
	/** The "average" difficulty. */
	A {
		@Override
		public String toString() {
			return A_TITLE;
		}
	},
	/** The "hard" difficulty. */
	H {
		@Override
		public String toString() {
			return H_TITLE;
		}
	},
	/** The "very hard" difficulty. */
	VH {
		@Override
		public String toString() {
			return VH_TITLE;
		}
	},
	/** The "wildcard" difficulty. */
	W {
		@Override
		public String toString() {
			return W_TITLE;
		}

		@Override
		public int getBaseRelativeLevel() {
			return VH.getBaseRelativeLevel();
		}
	};

	@Localize("E")
	@Localize(locale = "de", value = "E")
	@Localize(locale = "ru", value = "Л")
	static String	E_TITLE;
	@Localize("A")
	@Localize(locale = "de", value = "D")
	@Localize(locale = "ru", value = "С")
	static String	A_TITLE;
	@Localize("H")
	@Localize(locale = "de", value = "S")
	@Localize(locale = "ru", value = "Т")
	static String	H_TITLE;
	@Localize("VH")
	@Localize(locale = "de", value = "ES")
	@Localize(locale = "ru", value = "ОТ")
	static String	VH_TITLE;
	@Localize("W")
	@Localize(locale = "de", value = "W")
	@Localize(locale = "ru", value = "У")
	static String	W_TITLE;

	static {
		Localization.initialize();
	}

	/** @return The base relative skill level at 0 points. */
	public int getBaseRelativeLevel() {
		return -ordinal();
	}
}
