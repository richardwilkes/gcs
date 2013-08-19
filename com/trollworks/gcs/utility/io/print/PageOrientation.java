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

import java.awt.print.PageFormat;

import javax.print.attribute.standard.OrientationRequested;

/** Constants representing the various page orientation possibilities. */
public enum PageOrientation {
	/** Maps to {@link OrientationRequested#PORTRAIT}. */
	PORTRAIT {
		@Override public OrientationRequested getOrientationRequested() {
			return OrientationRequested.PORTRAIT;
		}

		@Override public String toString() {
			return MSG_PORTRAIT;
		}
	},
	/** Maps to {@link OrientationRequested#LANDSCAPE}. */
	LANDSCAPE {
		@Override public OrientationRequested getOrientationRequested() {
			return OrientationRequested.LANDSCAPE;
		}

		@Override public String toString() {
			return MSG_LANDSCAPE;
		}
	},
	/** Maps to {@link OrientationRequested#REVERSE_PORTRAIT}. */
	REVERSE_PORTRAIT {
		@Override public OrientationRequested getOrientationRequested() {
			return OrientationRequested.REVERSE_PORTRAIT;
		}

		@Override public String toString() {
			return MSG_REVERSE_PORTRAIT;
		}
	},
	/** Maps to {@link OrientationRequested#REVERSE_LANDSCAPE}. */
	REVERSE_LANDSCAPE {
		@Override public OrientationRequested getOrientationRequested() {
			return OrientationRequested.REVERSE_LANDSCAPE;
		}

		@Override public String toString() {
			return MSG_REVERSE_LANDSCAPE;
		}
	};

	static String	MSG_LANDSCAPE;
	static String	MSG_PORTRAIT;
	static String	MSG_REVERSE_LANDSCAPE;
	static String	MSG_REVERSE_PORTRAIT;

	static {
		LocalizedMessages.initialize(PageOrientation.class);
	}

	/** @return The orientation attribute. */
	public abstract OrientationRequested getOrientationRequested();

	/**
	 * @param orientation The {@link OrientationRequested} to load from.
	 * @return The page orientation.
	 */
	public static final PageOrientation get(OrientationRequested orientation) {
		for (PageOrientation one : values()) {
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
	public static final PageOrientation get(PageFormat format) {
		switch (format.getOrientation()) {
			case PageFormat.LANDSCAPE:
				return PageOrientation.LANDSCAPE;
			case PageFormat.REVERSE_LANDSCAPE:
				return PageOrientation.REVERSE_LANDSCAPE;
			case PageFormat.PORTRAIT:
			default:
				return PageOrientation.PORTRAIT;
		}
	}
}
