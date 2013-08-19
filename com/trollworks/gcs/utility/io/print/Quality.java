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

import javax.print.attribute.standard.PrintQuality;

/** Constants representing the various print quality possibilities. */
public enum Quality {
	/** Maps to {@link PrintQuality#HIGH}. */
	HIGH {
		@Override public PrintQuality getPrintQuality() {
			return PrintQuality.HIGH;
		}

		@Override public String toString() {
			return MSG_HIGH;
		}
	},
	/** Maps to {@link PrintQuality#NORMAL}. */
	NORMAL {
		@Override public PrintQuality getPrintQuality() {
			return PrintQuality.NORMAL;
		}

		@Override public String toString() {
			return MSG_NORMAL;
		}
	},
	/** Maps to {@link PrintQuality#DRAFT}. */
	DRAFT {
		@Override public PrintQuality getPrintQuality() {
			return PrintQuality.DRAFT;
		}

		@Override public String toString() {
			return MSG_DRAFT;
		}
	};

	static String	MSG_HIGH;
	static String	MSG_NORMAL;
	static String	MSG_DRAFT;

	static {
		LocalizedMessages.initialize(Quality.class);
	}

	/** @return The print quality attribute. */
	public abstract PrintQuality getPrintQuality();

	/**
	 * @param sides The {@link PrintQuality} to load from.
	 * @return The sides.
	 */
	public static final Quality get(PrintQuality sides) {
		for (Quality one : values()) {
			if (one.getPrintQuality() == sides) {
				return one;
			}
		}
		return NORMAL;
	}
}
