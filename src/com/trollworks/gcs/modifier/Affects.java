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

package com.trollworks.gcs.modifier;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;

/** Describes how a {@link Modifier} affects the point cost. */
public enum Affects {
	/** Affects the total cost. */
	TOTAL {
		@Override
		public String toString() {
			return TOTAL_TITLE;
		}

		@Override
		public String getShortTitle() {
			return TOTAL_SHORT;
		}
	},
	/** Affects only the base cost, not the leveled cost. */
	BASE_ONLY {
		@Override
		public String toString() {
			return BASE_ONLY_TITLE;
		}

		@Override
		public String getShortTitle() {
			return BASE_ONLY_SHORT;
		}
	},
	/** Affects only the leveled cost, not the base cost. */
	LEVELS_ONLY {
		@Override
		public String toString() {
			return LEVELS_ONLY_TITLE;
		}

		@Override
		public String getShortTitle() {
			return LEVELS_ONLY_SHORT;
		}
	};

	@Localize("to cost")
	static String	TOTAL_TITLE;
	@Localize("")
	static String	TOTAL_SHORT;
	@Localize("to base cost only")
	static String	BASE_ONLY_TITLE;
	@Localize("(base only)")
	static String	BASE_ONLY_SHORT;
	@Localize("to leveled cost only")
	static String	LEVELS_ONLY_TITLE;
	@Localize("(levels only)")
	static String	LEVELS_ONLY_SHORT;

	static {
		Localization.initialize();
	}

	/** @return The short version of the title. */
	public abstract String getShortTitle();
}
