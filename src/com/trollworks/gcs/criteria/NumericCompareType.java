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
import com.trollworks.toolkit.utility.Localization;

import java.text.MessageFormat;

/** The allowed numeric comparison types. */
public enum NumericCompareType {
	/** The comparison for "is". */
	IS {
		@Override
		public String toString() {
			return IS_TITLE;
		}

		@Override
		public String getDescription() {
			return IS_DESCRIPTION;
		}

		@Override
		String getDescriptionFormat() {
			return IS_FORMAT;
		}
	},
	/** The comparison for "is at least". */
	AT_LEAST {
		@Override
		public String toString() {
			return AT_LEAST_TITLE;
		}

		@Override
		public String getDescription() {
			return AT_LEAST_DESCRIPTION;
		}

		@Override
		String getDescriptionFormat() {
			return AT_LEAST_FORMAT;
		}
	},
	/** The comparison for "is at most". */
	AT_MOST {
		@Override
		public String toString() {
			return AT_MOST_TITLE;
		}

		@Override
		public String getDescription() {
			return AT_MOST_DESCRIPTION;
		}

		@Override
		String getDescriptionFormat() {
			return AT_MOST_FORMAT;
		}
	};

	@Localize("exactly")
	static String	IS_TITLE;
	@Localize("at least")
	static String	AT_LEAST_TITLE;
	@Localize("at most")
	static String	AT_MOST_TITLE;
	@Localize("{0}\"{1}\"")
	static String	IS_FORMAT;
	@Localize("{0}at least \"{1}\"")
	static String	AT_LEAST_FORMAT;
	@Localize("{0}at most \"{1}\"")
	static String	AT_MOST_FORMAT;
	@Localize("is")
	static String	IS_DESCRIPTION;
	@Localize("is at least")
	static String	AT_LEAST_DESCRIPTION;
	@Localize("is at most")
	static String	AT_MOST_DESCRIPTION;

	static {
		Localization.initialize();
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
