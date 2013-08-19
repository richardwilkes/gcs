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

package com.trollworks.gcs.advantage;

import com.trollworks.ttk.utility.LocalizedMessages;

/** The types of {@link Advantage} containers. */
public enum AdvantageContainerType {
	/** The standard grouping container type. */
	GROUP {
		@Override public String toString() {
			return MSG_GROUP;
		}
	},
	/**
	 * The meta-trait grouping container type. Acts as one normal trait, listed as an advantage if
	 * its point total is positive, or a disadvantage if it is negative.
	 */
	METATRAIT {
		@Override public String toString() {
			return MSG_METATRAIT;
		}
	},
	/**
	 * The race grouping container type. Its point cost is tracked separately from normal advantages
	 * and disadvantages.
	 */
	RACE {
		@Override public String toString() {
			return MSG_RACE;
		}
	},
	/**
	 * The alternative abilities grouping container type. It behaves similar to a {@link #METATRAIT},
	 * but applies the rules for alternative abilities (see B61 and P11) to its immediate children.
	 */
	ALTERNATIVE_ABILITIES {
		@Override public String toString() {
			return MSG_ALTERNATIVE_ABILITIES;
		}
	};

	static String	MSG_GROUP;
	static String	MSG_METATRAIT;
	static String	MSG_RACE;
	static String	MSG_ALTERNATIVE_ABILITIES;

	static {
		LocalizedMessages.initialize(AdvantageContainerType.class);
	}
}
