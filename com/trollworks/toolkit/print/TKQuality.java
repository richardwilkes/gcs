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

import javax.print.attribute.standard.PrintQuality;

/** Constants representing the various print quality possibilities. */
public enum TKQuality {
	/** Maps to {@link PrintQuality#HIGH}. */
	HIGH(Msgs.HIGH, PrintQuality.HIGH),
	/** Maps to {@link PrintQuality#NORMAL}. */
	NORMAL(Msgs.NORMAL, PrintQuality.NORMAL),
	/** Maps to {@link PrintQuality#DRAFT}. */
	DRAFT(Msgs.DRAFT, PrintQuality.DRAFT);

	private String			mName;
	private PrintQuality	mValue;

	private TKQuality(String name, PrintQuality value) {
		mName = name;
		mValue = value;
	}

	/** @return The print quality attribute. */
	public PrintQuality getPrintQuality() {
		return mValue;
	}

	@Override public String toString() {
		return mName;
	}

	/**
	 * @param sides The {@link PrintQuality} to load from.
	 * @return The sides.
	 */
	public static final TKQuality get(PrintQuality sides) {
		for (TKQuality one : values()) {
			if (one.getPrintQuality() == sides) {
				return one;
			}
		}
		return NORMAL;
	}
}
