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

import javax.print.attribute.standard.Sides;

/** Constants representing the various page side possibilities. */
public enum TKPageSides {
	/** Maps to {@link Sides#ONE_SIDED}. */
	SINGLE(Msgs.SINGLE, Sides.ONE_SIDED),
	/** Maps to {@link Sides#DUPLEX}. */
	DUPLEX(Msgs.DUPLEX, Sides.DUPLEX),
	/** Maps to {@link Sides#TUMBLE}. */
	TUMBLE(Msgs.TUMBLE, Sides.TUMBLE);

	private String	mName;
	private Sides	mValue;

	private TKPageSides(String name, Sides value) {
		mName = name;
		mValue = value;
	}

	/** @return The sides attribute. */
	public Sides getSides() {
		return mValue;
	}

	@Override public String toString() {
		return mName;
	}

	/**
	 * @param sides The {@link Sides} to load from.
	 * @return The sides.
	 */
	public static final TKPageSides get(Sides sides) {
		for (TKPageSides one : values()) {
			if (one.getSides() == sides) {
				return one;
			}
		}
		return SINGLE;
	}
}
