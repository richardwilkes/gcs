/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
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
	@Localize(locale = "de", value = "beliebig ist")
	@Localize(locale = "ru", value = "люб(ое,ая)")
	static String	IS_ANYTHING_TITLE;
	@Localize("is")
	@Localize(locale = "de", value = "lautet")
	@Localize(locale = "ru", value = " ")
	static String	IS_TITLE;
	@Localize("is not")
	@Localize(locale = "de", value = "nicht lautet")
	@Localize(locale = "ru", value = "не")
	static String	IS_NOT_TITLE;
	@Localize("contains")
	@Localize(locale = "de", value = "enthält")
	@Localize(locale = "ru", value = "содержит")
	static String	CONTAINS_TITLE;
	@Localize("does not contain")
	@Localize(locale = "de", value = "nicht enthält")
	@Localize(locale = "ru", value = "не содержит")
	static String	DOES_NOT_CONTAIN_TITLE;
	@Localize("starts with")
	@Localize(locale = "de", value = "beginnt mit")
	@Localize(locale = "ru", value = "начинается с")
	static String	STARTS_WITH_TITLE;
	@Localize("does not start with")
	@Localize(locale = "de", value = "nicht beginnt mit")
	@Localize(locale = "ru", value = "не начинается с")
	static String	DOES_NOT_START_WITH_TITLE;
	@Localize("ends with")
	@Localize(locale = "de", value = "endet auf")
	@Localize(locale = "ru", value = "заканчивается на")
	static String	ENDS_WITH_TITLE;
	@Localize("does not end with")
	@Localize(locale = "de", value = "nicht endet auf")
	@Localize(locale = "ru", value = "не заканчивается на")
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
}
