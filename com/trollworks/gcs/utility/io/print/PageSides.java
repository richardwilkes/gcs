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

import javax.print.attribute.standard.Sides;

/** Constants representing the various page side possibilities. */
public enum PageSides {
	/** Maps to {@link Sides#ONE_SIDED}. */
	SINGLE {
		@Override public Sides getSides() {
			return Sides.ONE_SIDED;
		}

		@Override public String toString() {
			return MSG_SINGLE;
		}
	},
	/** Maps to {@link Sides#DUPLEX}. */
	DUPLEX {
		@Override public Sides getSides() {
			return Sides.DUPLEX;
		}

		@Override public String toString() {
			return MSG_DUPLEX;
		}
	},
	/** Maps to {@link Sides#TUMBLE}. */
	TUMBLE {
		@Override public Sides getSides() {
			return Sides.TUMBLE;
		}

		@Override public String toString() {
			return MSG_TUMBLE;
		}
	};

	static String	MSG_SINGLE;
	static String	MSG_DUPLEX;
	static String	MSG_TUMBLE;

	static {
		LocalizedMessages.initialize(PageSides.class);
	}

	/** @return The sides attribute. */
	public abstract Sides getSides();

	/**
	 * @param sides The {@link Sides} to load from.
	 * @return The sides.
	 */
	public static final PageSides get(Sides sides) {
		for (PageSides one : values()) {
			if (one.getSides() == sides) {
				return one;
			}
		}
		return SINGLE;
	}
}
