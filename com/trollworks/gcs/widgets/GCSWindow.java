/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.ttk.widgets.AppWindow;

import java.awt.GraphicsConfiguration;
import java.awt.image.BufferedImage;

/** Provides a base OS-level window. */
public class GCSWindow extends AppWindow {
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
