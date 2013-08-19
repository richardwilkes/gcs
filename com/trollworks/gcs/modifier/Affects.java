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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.modifier;

import com.trollworks.ttk.utility.LocalizedMessages;

/** Describes how a {@link Modifier} affects the point cost. */
public enum Affects {
	/** Affects the total cost. */
	TOTAL {
		@Override public String toString() {
			return MSG_TOTAL;
		}

		@Override public String getShortTitle() {
			return MSG_TOTAL_SHORT;
		}
	},
	/** Affects only the base cost, not the leveled cost. */
	BASE_ONLY {
		@Override public String toString() {
			return MSG_BASE_ONLY;
		}

		@Override public String getShortTitle() {
			return MSG_BASE_ONLY_SHORT;
		}
	},
	/** Affects only the leveled cost, not the base cost. */
	LEVELS_ONLY {
		@Override public String toString() {
			return MSG_LEVELS_ONLY;
		}

		@Override public String getShortTitle() {
			return MSG_LEVELS_ONLY_SHORT;
		}
	};

	static String	MSG_TOTAL;
	static String	MSG_TOTAL_SHORT;
	static String	MSG_BASE_ONLY;
	static String	MSG_BASE_ONLY_SHORT;
	static String	MSG_LEVELS_ONLY;
	static String	MSG_LEVELS_ONLY_SHORT;

	static {
		LocalizedMessages.initialize(Affects.class);
	}

	/** @return The short version of the title. */
	public abstract String getShortTitle();
}
