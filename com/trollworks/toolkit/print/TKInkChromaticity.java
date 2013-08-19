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

package com.trollworks.toolkit.print;

import javax.print.attribute.standard.Chromaticity;

/** Constants representing the various ink chromaticity possibilities. */
public enum TKInkChromaticity {
	/** Maps to {@link Chromaticity#COLOR}. */
	COLOR(Msgs.COLOR, Chromaticity.COLOR),
	/** Maps to {@link Chromaticity#MONOCHROME}. */
	MONOCHROME(Msgs.MONOCHROME, Chromaticity.MONOCHROME);

	private String			mName;
	private Chromaticity	mValue;

	private TKInkChromaticity(String name, Chromaticity value) {
		mName = name;
		mValue = value;
	}

	/** @return The chromaticity attribute. */
	public Chromaticity getChromaticity() {
		return mValue;
	}

	@Override public String toString() {
		return mName;
	}

	/**
	 * @param chromaticity The {@link Chromaticity} to load from.
	 * @return The chromaticity.
	 */
	public static final TKInkChromaticity get(Chromaticity chromaticity) {
		for (TKInkChromaticity one : values()) {
			if (one.getChromaticity() == chromaticity) {
				return one;
			}
		}
		return COLOR;
	}
}
