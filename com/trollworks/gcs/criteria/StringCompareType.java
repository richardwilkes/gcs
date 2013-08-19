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

import com.trollworks.ttk.collections.Enums;
import com.trollworks.ttk.utility.LocalizedMessages;

/** The allowed string comparison types. */
public enum StringCompareType {
	/** The comparison for "is anything". */
	IS_ANYTHING {
		@Override
		public String toString() {
			return MSG_IS_ANYTHING;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return true;
		}
	},
	/** The comparison for "is". */
	IS {
		@Override
		public String toString() {
			return MSG_IS;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return data.equalsIgnoreCase(qualifier);
		}
	},
	/** The comparison for "is not". */
	IS_NOT {
		@Override
		public String toString() {
			return MSG_IS_NOT;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return !data.equalsIgnoreCase(qualifier);
		}
	},
	/** The comparison for "contains". */
	CONTAINS {
		@Override
		public String toString() {
			return MSG_CONTAINS;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return data.toLowerCase().indexOf(qualifier.toLowerCase()) != -1;
		}
	},
	/** The comparison for "does not contain". */
	DOES_NOT_CONTAIN {
		@Override
		public String toString() {
			return MSG_DOES_NOT_CONTAIN;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return data.toLowerCase().indexOf(qualifier.toLowerCase()) == -1;
		}
	},
	/** The comparison for "starts with". */
	STARTS_WITH {
		@Override
		public String toString() {
			return MSG_STARTS_WITH;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return data.toLowerCase().startsWith(qualifier.toLowerCase());
		}
	},
	/** The comparison for "does not start with". */
	DOES_NOT_START_WITH {
		@Override
		public String toString() {
			return MSG_DOES_NOT_START_WITH;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return !data.toLowerCase().startsWith(qualifier.toLowerCase());
		}
	},
	/** The comparison for "ends with". */
	ENDS_WITH {
		@Override
		public String toString() {
			return MSG_ENDS_WITH;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return data.toLowerCase().endsWith(qualifier.toLowerCase());
		}
	},
	/** The comparison for "does not end with". */
	DOES_NOT_END_WITH {
		@Override
		public String toString() {
			return MSG_DOES_NOT_END_WITH;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return !data.toLowerCase().endsWith(qualifier.toLowerCase());
		}
	};

	static String	MSG_IS;
	static String	MSG_IS_ANYTHING;
	static String	MSG_IS_NOT;
	static String	MSG_CONTAINS;
	static String	MSG_DOES_NOT_CONTAIN;
	static String	MSG_STARTS_WITH;
	static String	MSG_DOES_NOT_START_WITH;
	static String	MSG_ENDS_WITH;
	static String	MSG_DOES_NOT_END_WITH;

	static {
		LocalizedMessages.initialize(StringCompareType.class);
	}

	/**
	 * @param qualifier The qualifier.
	 * @return The description of this comparison type.
	 */
	public String describe(String qualifier) {
		StringBuilder builder = new StringBuilder();
		builder.append(toString());
		builder.append(" \""); //$NON-NLS-1$
		builder.append(qualifier);
		builder.append('"');
		return builder.toString();
	}

	/**
	 * Performs a comparison.
	 * 
	 * @param qualifier The qualifier to use in conjuction with this {@link StringCompareType}.
	 * @param data The data to check.
	 * @return Whether the data matches the criteria or not.
	 */
	public abstract boolean matches(String qualifier, String data);

	/**
	 * @param buffer The buffer to load from.
	 * @return The units representing the buffer's description.
	 */
	public static final StringCompareType get(String buffer) {
		StringCompareType result = Enums.extract(buffer, values());
		if (result == null) {
			// Check a few others, for legacy reasons
			if ("starts".equalsIgnoreCase(buffer)) { //$NON-NLS-1$
				result = STARTS_WITH;
			} else if ("ends".equalsIgnoreCase(buffer)) { //$NON-NLS-1$
				result = ENDS_WITH;
			} else {
				result = IS;
			}
		}
		return result;
	}
}
