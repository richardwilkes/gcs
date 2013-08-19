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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets;

import java.awt.Dimension;
import java.awt.Window;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;

/** Ensures a window is never resized below its minimum size or above its maximum size settings. */
public class WindowSizeEnforcer implements ComponentListener {
	private Window	mWindow;

	/**
	 * Creates a new {@link WindowSizeEnforcer}.
	 * 
	 * @param window The window to monitor.
	 */
	public WindowSizeEnforcer(Window window) {
		mWindow = window;
		mWindow.addComponentListener(this);
	}

	public void componentHidden(ComponentEvent event) {
		// Not used.
	}

	public void componentMoved(ComponentEvent event) {
		// Not used.
	}

	public void componentResized(ComponentEvent event) {
		Dimension origSize = mWindow.getSize();
		Dimension otherSize = mWindow.getMinimumSize();
		int width = origSize.width;
		int height = origSize.height;

		if (width < otherSize.width) {
			width = otherSize.width;
		}
		if (height < otherSize.height) {
			height = otherSize.height;
		}
		otherSize = mWindow.getMaximumSize();
		if (width > otherSize.width) {
			width = otherSize.width;
		}
		if (height > otherSize.height) {
			height = otherSize.height;
		}
		if (width != origSize.width || height != origSize.height) {
			mWindow.setSize(width, height);
		}
	}

	public void componentShown(ComponentEvent event) {
		// Not used.
	}
}
