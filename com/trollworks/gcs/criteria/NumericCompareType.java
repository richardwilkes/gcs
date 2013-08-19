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

package com.trollworks.gcs.criteria;

import com.trollworks.ttk.utility.LocalizedMessages;

import java.text.MessageFormat;

/** The allowed numeric comparison types. */
public enum NumericCompareType {
	/** The comparison for "is". */
	IS {
		@Override public String toString() {
			return MSG_IS;
		}

		@Override public String getDescription() {
			return MSG_IS_DESCRIPTION;
		}

		@Override String getDescriptionFormat() {
			return MSG_IS_FORMAT;
		}
	},
	/** The comparison for "is at least". */
	AT_LEAST {
		@Override public String toString() {
			return MSG_AT_LEAST;
		}

		@Override public String getDescription() {
			return MSG_AT_LEAST_DESCRIPTION;
		}

		@Override String getDescriptionFormat() {
			return MSG_AT_LEAST_FORMAT;
		}
	},
	/** The comparison for "is at most". */
	AT_MOST {
		@Override public String toString() {
			return MSG_AT_MOST;
		}

		@Override public String getDescription() {
			return MSG_AT_MOST_DESCRIPTION;
		}

		@Override String getDescriptionFormat() {
			return MSG_AT_MOST_FORMAT;
		}
	};

	static String	MSG_IS;
	static String	MSG_AT_LEAST;
	static String	MSG_AT_MOST;
	static String	MSG_IS_FORMAT;
	static String	MSG_AT_LEAST_FORMAT;
	static String	MSG_AT_MOST_FORMAT;
	static String	MSG_IS_DESCRIPTION;
	static String	MSG_AT_LEAST_DESCRIPTION;
	static String	MSG_AT_MOST_DESCRIPTION;

	static {
		LocalizedMessages.initialize(NumericCompareType.class);
	}

	/** @return A description of this object. */
	public abstract String getDescription();

	abstract String getDescriptionFormat();

	/**
	 * @param prefix A prefix to place before the description.
	 * @param qualifier The qualifier to use.
	 * @return A formatted description of this object.
	 */
	public String format(String prefix, String qualifier) {
		return MessageFormat.format(getDescriptionFormat(), prefix, qualifier);
	}
}
