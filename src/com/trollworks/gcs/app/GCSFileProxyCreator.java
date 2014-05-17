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

package com.trollworks.gcs.app;

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.toolkit.ui.menu.file.FileProxy;
import com.trollworks.toolkit.ui.menu.file.FileProxyCreator;
import com.trollworks.toolkit.ui.widget.BaseWindow;
import com.trollworks.toolkit.ui.widget.dock.Dock;
import com.trollworks.toolkit.ui.widget.dock.Dockable;

import java.awt.KeyboardFocusManager;
import java.awt.Window;
import java.io.File;
import java.io.IOException;
import java.util.ArrayList;

/**
 * Responsible for creating new {@link Dockable}s and installing them into the {@link MainWindow}.
 */
public class GCSFileProxyCreator implements FileProxyCreator {
	@Override
	public FileProxy create(File file) throws IOException {
		Window window = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusedWindow();
		if (window == null || !(window instanceof MainWindow)) {
			ArrayList<MainWindow> windows = BaseWindow.getWindows(MainWindow.class);
			if (!windows.isEmpty()) {
				window = windows.get(0);
			}
		}
		if (window == null) {
			window = new MainWindow();
			window.pack();
			window.setVisible(true);
		}
		Dock dock = ((MainWindow) window).getDock();
		for (Dockable dockable : dock.getDockables()) {
			if (dockable instanceof LibraryExplorerDockable) {
				FileProxy proxy = ((LibraryExplorerDockable) dockable).open(file.toPath());
				proxy.toFrontAndFocus();
				return proxy;
			}
		}
		return null;
	}
}
