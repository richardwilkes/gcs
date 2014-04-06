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

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.toolkit.ui.menu.file.SignificantFrame;
import com.trollworks.toolkit.ui.widget.AppWindow;

import java.awt.GraphicsConfiguration;
import java.awt.image.BufferedImage;

/** Provides a base OS-level window. */
public class GCSWindow extends AppWindow implements SignificantFrame {
	/**
	 * Creates a new {@link AppWindow}.
	 * 
	 * @param title The window title. May be <code>null</code>.
	 * @param largeIcon The 32x32 window icon. OK to pass in a 16x16 icon here.
	 * @param smallIcon The 16x16 window icon.
	 */
	public GCSWindow(String title, BufferedImage largeIcon, BufferedImage smallIcon) {
		super(title, largeIcon, smallIcon);
	}

	/**
	 * Creates a new {@link AppWindow}.
	 * 
	 * @param title The title of the window.
	 * @param largeIcon The 32x32 window icon. OK to pass in a 16x16 icon here.
	 * @param smallIcon The 16x16 window icon.
	 * @param gc The graphics configuration to use.
	 * @param undecorated Whether to create an undecorated window, without menus.
	 */
	public GCSWindow(String title, BufferedImage largeIcon, BufferedImage smallIcon, GraphicsConfiguration gc, boolean undecorated) {
		super(title, largeIcon, smallIcon, gc, undecorated);
	}

	@Override
	public void setTitle(String title) {
		super.setTitle(DataMenu.filterTitle(title));
	}
}
