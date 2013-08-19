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

package com.trollworks.gcs.modifier;

import com.trollworks.ttk.utility.LocalizedMessages;

/** Describes how a {@link Modifier}'s point cost is applied. */
public enum CostType {
	/** Adds to the percentage multiplier. */
	PERCENTAGE {
		@Override public String toString() {
			return MSG_PERCENTAGE;
		}
	},
	/** Adds a constant to the base value prior to any multiplier or percentage adjustment. */
	POINTS {
		@Override public String toString() {
			return MSG_POINTS;
		}
	},
	/** Multiplies the final cost by a constant. */
	MULTIPLIER {
		@Override public String toString() {
			return MSG_MULTIPLIER;
		}
	};

	static String	MSG_PERCENTAGE;
	static String	MSG_POINTS;
	static String	MSG_MULTIPLIER;

	static {
		LocalizedMessages.initialize(CostType.class);
	}
}
