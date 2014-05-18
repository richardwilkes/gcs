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
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.menu.edit.Undoable;
import com.trollworks.toolkit.ui.menu.file.SignificantFrame;
import com.trollworks.toolkit.ui.widget.AppWindow;
import com.trollworks.toolkit.ui.widget.BaseWindow;
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.dock.Dock;
import com.trollworks.toolkit.ui.widget.dock.DockContainer;
import com.trollworks.toolkit.ui.widget.dock.DockLocation;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.utility.Geometry;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.KeyboardFocusManager;
import java.awt.Rectangle;
import java.awt.Window;
import java.util.ArrayList;

/** The main GCS window. */
public class MainWindow extends AppWindow implements SignificantFrame, JumpToSearchTarget {
	@Localize("Workspace")
	private static String	TITLE;

	static {
		Localization.initialize();
	}

	private Toolbar			mToolbar;
	private Dock			mDock;

	/** @return The {@link MainWindow}. */
	public static MainWindow get() {
		Window window = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusedWindow();
		if (window == null || !(window instanceof MainWindow)) {
			ArrayList<MainWindow> windows = BaseWindow.getWindows(MainWindow.class);
			if (!windows.isEmpty()) {
				window = windows.get(0);
			}
		}
		if (window == null) {
			window = new MainWindow();
		}
		return (MainWindow) window;
	}

	private MainWindow() {
		super(TITLE, GCSImages.getAppIcons());
		Container content = getContentPane();
		mToolbar = new Toolbar();
		content.add(mToolbar, BorderLayout.NORTH);
		mDock = new Dock();
		content.add(mDock, BorderLayout.CENTER);
		LibraryExplorerDockable libraryExplorer = new LibraryExplorerDockable();
		mDock.dock(libraryExplorer, DockLocation.WEST);
		mDock.getLayout().findLayout(libraryExplorer.getDockContainer()).setDividerPosition(200);
		restoreBounds();
		setVisible(true);
	}

	@Override
	public void pack() {
		super.pack();
		setBounds(Geometry.inset(20, new Rectangle(GraphicsUtilities.getMaximumWindowBounds())));
	}

	@Override
	public String getWindowPrefsPrefix() {
		return "Workspace."; //$NON-NLS-1$
	}

	/** @return The {@link Dock}. */
	public Dock getDock() {
		return mDock;
	}

	@Override
	public StdUndoManager getUndoManager() {
		DockContainer dc = mDock.getFocusedDockContainer();
		if (dc != null) {
			Dockable dockable = dc.getCurrentDockable();
			if (dockable instanceof Undoable) {
				return ((Undoable) dockable).getUndoManager();
			}
		}
		return super.getUndoManager();
	}

	@Override
	public boolean isJumpToSearchAvailable() {
		DockContainer dc = mDock.getFocusedDockContainer();
		if (dc != null) {
			Dockable dockable = dc.getCurrentDockable();
			if (dockable instanceof JumpToSearchTarget) {
				return ((JumpToSearchTarget) dockable).isJumpToSearchAvailable();
			}
		}
		return false;
	}

	@Override
	public void jumpToSearchField() {
		DockContainer dc = mDock.getFocusedDockContainer();
		if (dc != null) {
			Dockable dockable = dc.getCurrentDockable();
			if (dockable instanceof JumpToSearchTarget) {
				((JumpToSearchTarget) dockable).jumpToSearchField();
			}
		}
	}
}
