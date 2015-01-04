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

package com.trollworks.gcs.library;

import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.tree.TreeContainerRow;
import com.trollworks.toolkit.ui.widget.tree.TreeRow;

/** A {@link TreeRow} that represents a directory in the library explorer. */
public class LibraryDirectoryRow extends TreeContainerRow implements LibraryExplorerRow {
	private String	mName;

	/** @param name The name of the directory. */
	public LibraryDirectoryRow(String name) {
		mName = name;
	}

	@Override
	public String getSelectionKey() {
		TreeContainerRow parent = getParent();
		return parent instanceof LibraryDirectoryRow ? ((LibraryDirectoryRow) parent).getSelectionKey() + "/" + mName : mName; //$NON-NLS-1$
	}

	@Override
	public StdImage getIcon() {
		return StdImage.FOLDER.getImage(16);
	}

	@Override
	public String getName() {
		return mName;
	}
}
