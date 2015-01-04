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

package com.trollworks.gcs.advantage;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Enums;
import com.trollworks.toolkit.utility.text.Numbers;

/** The possible self-control rolls, from page B121. */
public enum SelfControlRoll {
	/** Rarely. */
	CR6 {
		@Override
		public String toString() {
			return CR6_TITLE;
		}

		@Override
		public double getMultiplier() {
			return 2;
		}

		@Override
		public int getCR() {
			return 6;
		}
	},
	/** Fairly often. */
	CR9 {
		@Override
		public String toString() {
			return CR9_TITLE;
		}

		@Override
		public double getMultiplier() {
			return 1.5;
		}

		@Override
		public int getCR() {
			return 9;
		}
	},
	/** Quite often. */
	CR12 {
		@Override
		public String toString() {
			return CR12_TITLE;
		}

		@Override
		public double getMultiplier() {
			return 1;
		}

		@Override
		public int getCR() {
			return 12;
		}
	},
	/** Almost all the time. */
	CR15 {
		@Override
		public String toString() {
			return CR15_TITLE;
		}

		@Override
		public double getMultiplier() {
			return 0.5;
		}

		@Override
		public int getCR() {
			return 15;
		}
	},
	/** No self-control roll. */
	NONE_REQUIRED {
		@Override
		public String toString() {
			return NONE_REQUIRED_TITLE;
		}

		@Override
		public double getMultiplier() {
			return 1;
		}

		@Override
		public int getCR() {
			return Integer.MAX_VALUE;
		}

		@Override
		public String getDescriptionWithCost() {
			return ""; //$NON-NLS-1$
		}

		@Override
		public void save(XMLWriter out, String tag, SelfControlRollAdjustments adj) {
			// Do nothing.
		}
	};

	@Localize("CR: 6 (Rarely)")
	@Localize(locale = "de", value = "SBP: 6 (selten)")
	@Localize(locale = "ru", value = "СК: 6 (редко)")
	static String				CR6_TITLE;
	@Localize("CR: 9 (Fairly Often)")
	@Localize(locale = "de", value = "SBP: 9 (öfters)")
	@Localize(locale = "ru", value = "СК: 9 (часто)")
	static String				CR9_TITLE;
	@Localize("CR: 12 (Quite Often)")
	@Localize(locale = "de", value = "SBP: 12 (häufig)")
	@Localize(locale = "ru", value = "СК: 12 (достаточно часто)")
	static String				CR12_TITLE;
	@Localize("CR: 15 (Almost All The Time)")
	@Localize(locale = "de", value = "SBP: 15 (fast immer)")
	@Localize(locale = "ru", value = "СК: 15 (почти всегда)")
	static String				CR15_TITLE;
	@Localize("None Required")
	@Localize(locale = "de", value = "Keine benötigt")
	@Localize(locale = "ru", value = "Не требуется")
	static String				NONE_REQUIRED_TITLE;

	static {
		Localization.initialize();
	}

	/** The attribute tag use for {@link SelfControlRollAdjustments}. */
	public static final String	ATTR_ADJUSTMENT	= "adj";	//$NON-NLS-1$

	/**
	 * @param tagValue The value within a tag representing a {@link SelfControlRoll}.
	 * @return The actual {@link SelfControlRoll}.
	 */
	public static final SelfControlRoll get(String tagValue) {
		int value = Numbers.getInteger(tagValue, Integer.MAX_VALUE);
		for (SelfControlRoll cr : values()) {
			if (cr.getCR() == value) {
				return cr;
			}
		}
		return NONE_REQUIRED;
	}

	/** @return The description, along with the cost. */
	public String getDescriptionWithCost() {
		return toString() + ", x" + getMultiplier(); //$NON-NLS-1$
	}

	/** @return The cost multiplier. */
	public abstract double getMultiplier();

	/** @return The minimum number to roll to retain control. */
	public abstract int getCR();

	/**
	 * @param out The {@link XMLWriter} to use.
	 * @param tag The XML tag to use.
	 * @param adj The {@link SelfControlRollAdjustments} being used.
	 */
	public void save(XMLWriter out, String tag, SelfControlRollAdjustments adj) {
		if (adj != SelfControlRollAdjustments.NONE) {
			out.simpleTagWithAttribute(tag, getCR(), ATTR_ADJUSTMENT, Enums.toId(adj));
		} else {
			out.simpleTag(tag, getCR());
		}
	}
}
