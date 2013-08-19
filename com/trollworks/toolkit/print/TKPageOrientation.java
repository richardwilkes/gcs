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

import java.awt.print.PageFormat;

import javax.print.attribute.standard.OrientationRequested;

/** Constants representing the various page orientation possibilities. */
public enum TKPageOrientation {
	/** Maps to {@link OrientationRequested#PORTRAIT}. */
	PORTRAIT(Msgs.PORTRAIT, OrientationRequested.PORTRAIT),
	/** Maps to {@link OrientationRequested#LANDSCAPE}. */
	LANDSCAPE(Msgs.LANDSCAPE, OrientationRequested.LANDSCAPE),
	/** Maps to {@link OrientationRequested#REVERSE_PORTRAIT}. */
	REVERSE_PORTRAIT(Msgs.REVERSE_PORTRAIT, OrientationRequested.REVERSE_PORTRAIT),
	/** Maps to {@link OrientationRequested#REVERSE_LANDSCAPE}. */
	REVERSE_LANDSCAPE(Msgs.REVERSE_LANDSCAPE, OrientationRequested.REVERSE_LANDSCAPE);

	private String					mName;
	private OrientationRequested	mValue;

	private TKPageOrientation(String name, OrientationRequested value) {
		mName = name;
		mValue = value;
	}

	/** @return The orientation attribute. */
	public OrientationRequested getOrientationRequested() {
		return mValue;
	}

	@Override public String toString() {
		return mName;
	}

	/**
	 * @param orientation The {@link OrientationRequested} to load from.
	 * @return The page orientation.
	 */
	public static final TKPageOrientation get(OrientationRequested orientation) {
		for (TKPageOrientation one : values()) {
			if (one.getOrientationRequested() == orientation) {
				return one;
			}
		}
		return PORTRAIT;
	}

	/**
	 * @param format The {@link PageFormat} to load from.
	 * @return The page orientation.
	 */
	public static final TKPageOrientation get(PageFormat format) {
		switch (format.getOrientation()) {
			case PageFormat.LANDSCAPE:
				return TKPageOrientation.LANDSCAPE;
			case PageFormat.REVERSE_LANDSCAPE:
				return TKPageOrientation.REVERSE_LANDSCAPE;
			case PageFormat.PORTRAIT:
			default:
				return TKPageOrientation.PORTRAIT;
		}
	}
}
