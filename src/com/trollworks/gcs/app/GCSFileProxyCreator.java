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

package com.trollworks.gcs.app;

import com.trollworks.gcs.common.Workspace;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.utility.FileProxy;
import com.trollworks.toolkit.utility.FileProxyCreator;

import java.io.File;
import java.io.IOException;

/**
 * Responsible for creating new {@link Dockable}s and installing them into the {@link Workspace}.
 */
public class GCSFileProxyCreator implements FileProxyCreator {
	@Override
	public FileProxy create(File file) throws IOException {
		LibraryExplorerDockable library = LibraryExplorerDockable.get();
		if (library != null) {
			FileProxy proxy = library.open(file.toPath());
			proxy.toFrontAndFocus();
			return proxy;
		}
		return null;
	}
}
