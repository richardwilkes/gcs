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

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.GraphicsUtilities;
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

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.KeyboardFocusManager;
import java.awt.Rectangle;
import java.awt.Window;
import java.util.ArrayList;

/** The workspace, where all files can be viewed and edited. */
public class Workspace extends AppWindow implements SignificantFrame, JumpToSearchTarget {
	@Localize("GURPS Workspace")
	@Localize(locale = "de", value = "GURPS Charakter-Editor")
	@Localize(locale = "ru", value = "GURPS рабочее пространство")
	private static String	TITLE;

	static {
		Localization.initialize();
	}

	private Toolbar			mToolbar;
	private Dock			mDock;

	/** @return The {@link Workspace}. */
	public static Workspace get() {
		Window window = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusedWindow();
		if (window == null || !(window instanceof Workspace)) {
			ArrayList<Workspace> windows = BaseWindow.getWindows(Workspace.class);
			if (!windows.isEmpty()) {
				window = windows.get(0);
			}
		}
		if (window == null) {
			window = new Workspace();
		}
		return (Workspace) window;
	}

	private Workspace() {
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
		return "workspace."; //$NON-NLS-1$
	}

	/** @return The {@link Dock}. */
	public Dock getDock() {
		return mDock;
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
