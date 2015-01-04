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

package com.trollworks.gcs.menu.edit;

/** Objects which contain a search field should implement this method. */
public interface JumpToSearchTarget {
	/**
	 * @return <code>true</code> if issuing the {@link #jumpToSearchField()} command will do
	 *         anything.
	 */
	boolean isJumpToSearchAvailable();

	/** Cause the search field to become focused. */
	void jumpToSearchField();
}
