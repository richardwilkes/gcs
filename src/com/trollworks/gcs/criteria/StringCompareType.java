/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.criteria;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.collections.Enums;
import com.trollworks.toolkit.utility.Localization;

/** The allowed string comparison types. */
public enum StringCompareType {
	/** The comparison for "is anything". */
	IS_ANYTHING {
		@Override
		public String toString() {
			return IS_ANYTHING_TITLE;
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
			return IS_TITLE;
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
			return IS_NOT_TITLE;
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
			return CONTAINS_TITLE;
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
			return DOES_NOT_CONTAIN_TITLE;
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
			return STARTS_WITH_TITLE;
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
			return DOES_NOT_START_WITH_TITLE;
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
			return ENDS_WITH_TITLE;
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
			return DOES_NOT_END_WITH_TITLE;
		}

		@Override
		public boolean matches(String qualifier, String data) {
			return !data.toLowerCase().endsWith(qualifier.toLowerCase());
		}
	};

	@Localize("is anything")
	static String	IS_ANYTHING_TITLE;
	@Localize("is")
	static String	IS_TITLE;
	@Localize("is not")
	static String	IS_NOT_TITLE;
	@Localize("contains")
	static String	CONTAINS_TITLE;
	@Localize("does not contain")
	static String	DOES_NOT_CONTAIN_TITLE;
	@Localize("starts with")
	static String	STARTS_WITH_TITLE;
	@Localize("does not start with")
	static String	DOES_NOT_START_WITH_TITLE;
	@Localize("ends with")
	static String	ENDS_WITH_TITLE;
	@Localize("does not end with")
	static String	DOES_NOT_END_WITH_TITLE;

	static {
		Localization.initialize();
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
