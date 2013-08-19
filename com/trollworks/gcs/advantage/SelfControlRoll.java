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

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.text.NumberUtils;

/** The possible self-control rolls, from page B121. */
public enum SelfControlRoll {
	/** Rarely. */
	CR6 {
		@Override public double getMultiplier() {
			return 2;
		}

		@Override public String toString() {
			return MSG_CR6;
		}

		@Override public int getCR() {
			return 6;
		}
	},
	/** Fairly often. */
	CR9 {
		@Override public double getMultiplier() {
			return 1.5;
		}

		@Override public String toString() {
			return MSG_CR9;
		}

		@Override public int getCR() {
			return 9;
		}
	},
	/** Quite often. */
	CR12 {
		@Override public double getMultiplier() {
			return 1;
		}

		@Override public String toString() {
			return MSG_CR12;
		}

		@Override public int getCR() {
			return 12;
		}
	},
	/** Almost all the time. */
	CR15 {
		@Override public double getMultiplier() {
			return 0.5;
		}

		@Override public String toString() {
			return MSG_CR15;
		}

		@Override public int getCR() {
			return 15;
		}
	},
	/** No self-control roll. */
	NONE_REQUIRED {
		@Override public double getMultiplier() {
			return 1;
		}

		@Override public String toString() {
			return MSG_NONE;
		}

		@Override public int getCR() {
			return Integer.MAX_VALUE;
		}

		@Override public String getDescriptionWithCost() {
			return ""; //$NON-NLS-1$
		}

		@Override public void save(XMLWriter out, String tag, SelfControlRollAdjustments adj) {
			// Do nothing.
		}
	};

	static String				MSG_CR6;
	static String				MSG_CR9;
	static String				MSG_CR12;
	static String				MSG_CR15;
	static String				MSG_NONE;
	/** The attribute tag use for {@link SelfControlRollAdjustments}. */
	public static final String	ATTR_ADJUSTMENT	= "adj";	//$NON-NLS-1$

	static {
		LocalizedMessages.initialize(SelfControlRoll.class);
	}

	/**
	 * @param tagValue The value within a tag representing a {@link SelfControlRoll}.
	 * @return The actual {@link SelfControlRoll}.
	 */
	public static final SelfControlRoll get(String tagValue) {
		int value = NumberUtils.getNonLocalizedInteger(tagValue, Integer.MAX_VALUE);
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
			out.simpleTagWithAttribute(tag, getCR(), ATTR_ADJUSTMENT, adj.name().toLowerCase());
		} else {
			out.simpleTag(tag, getCR());
		}
	}
}
