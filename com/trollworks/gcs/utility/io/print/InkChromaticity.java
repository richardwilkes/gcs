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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.utility.io.print;

import com.trollworks.gcs.utility.io.LocalizedMessages;

import javax.print.attribute.standard.Chromaticity;

/** Constants representing the various ink chromaticity possibilities. */
public enum InkChromaticity {
	/** Maps to {@link Chromaticity#COLOR}. */
	COLOR {
		@Override public Chromaticity getChromaticity() {
			return Chromaticity.COLOR;
		}

		@Override public String toString() {
			return MSG_COLOR;
		}
	},
	/** Maps to {@link Chromaticity#MONOCHROME}. */
	MONOCHROME {
		@Override public Chromaticity getChromaticity() {
			return Chromaticity.MONOCHROME;
		}

		@Override public String toString() {
			return MSG_MONOCHROME;
		}
	};

	static String	MSG_COLOR;
	static String	MSG_MONOCHROME;

	static {
		LocalizedMessages.initialize(InkChromaticity.class);
	}

	/** @return The chromaticity attribute. */
	public abstract Chromaticity getChromaticity();

	/**
	 * @param chromaticity The {@link Chromaticity} to load from.
	 * @return The chromaticity.
	 */
	public static final InkChromaticity get(Chromaticity chromaticity) {
		for (InkChromaticity one : values()) {
			if (one.getChromaticity() == chromaticity) {
				return one;
			}
		}
		return MONOCHROME;
	}
}
