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

package com.trollworks.gcs.common;

import java.nio.file.Path;
import java.util.List;

/**
 * Objects that want to be called when the {@link ListCollectionThread} has new data must implement
 * this interface.
 */
public interface ListCollectionListener {
	/**
	 * Called whenever the {@link ListCollectionThread} has new data.
	 *
	 * @param lists The data. The first object in the list (and any sub-lists) is always a
	 *            {@link String} representing the name of a directory. Each subsequent item in the
	 *            list is either another {@link List} or a {@link Path}.
	 */
	void dataFileListUpdated(List<Object> lists);
}
