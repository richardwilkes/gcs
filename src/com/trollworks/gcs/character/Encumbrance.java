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

package com.trollworks.gcs.character;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** Valid encumbrance levels. */
public enum Encumbrance {
	NONE(1) {
		@Override
		public String toString() {
			return NONE_TITLE;
		}
	},
	LIGHT(2) {
		@Override
		public String toString() {
			return LIGHT_TITLE;
		}
	},
	MEDIUM(3) {
		@Override
		public String toString() {
			return MEDIUM_TITLE;
		}
	},
	HEAVY(6) {
		@Override
		public String toString() {
			return HEAVY_TITLE;
		}
	},
	EXTRA_HEAVY(10) {
		@Override
		public String toString() {
			return EXTRA_HEAVY_TITLE;
		}
	};

	@Localize("None")
	@Localize(locale = "de", value = "Keine")
	@Localize(locale = "ru", value = "Нет")
	protected static String	NONE_TITLE;
	@Localize("Light")
	@Localize(locale = "de", value = "Leicht")
	@Localize(locale = "ru", value = "Лёгкая")
	protected static String	LIGHT_TITLE;
	@Localize("Medium")
	@Localize(locale = "de", value = "Mittel")
	@Localize(locale = "ru", value = "Средние")
	protected static String	MEDIUM_TITLE;
	@Localize("Heavy")
	@Localize(locale = "de", value = "Schwer")
	@Localize(locale = "ru", value = "Тяжелый")
	protected static String	HEAVY_TITLE;
	@Localize("X-Heavy")
	@Localize(locale = "de", value = "Extrem")
	@Localize(locale = "ru", value = "Сверхтяж")
	protected static String	EXTRA_HEAVY_TITLE;

	static {
		Localization.initialize();
	}

	private int				mMultiplier;

	private Encumbrance(int multiplier) {
		mMultiplier = multiplier;
	}

	/** @return The weight multiplier associated with this level of encumbrance. */
	public int getWeightMultiplier() {
		return mMultiplier;
	}

	/** @return The penalty associated with this level of encumbrance. */
	public int getEncumbrancePenalty() {
		return -ordinal();
	}
}
