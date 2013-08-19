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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKAppWithUI;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;

import java.awt.Dimension;
import java.awt.Rectangle;
import java.awt.image.BufferedImage;

/** A standard about box window. */
public class TKAboutWindow extends TKWindow {
	/** Create a new, default about box. */
	public TKAboutWindow() {
		this(getDefaultPanel());
	}

	/**
	 * Create a new about box.
	 * 
	 * @param windowIcon The icon to use for the window.
	 * @param windowTitle The window title.
	 * @param graphic The primary graphic to display.
	 */
	public TKAboutWindow(String windowTitle, BufferedImage windowIcon, BufferedImage graphic) {
		this(windowTitle, windowIcon, new TKLabel(graphic));
	}

	/**
	 * Create a new about box.
	 * 
	 * @param panel The primary panel to display.
	 */
	public TKAboutWindow(TKPanel panel) {
		this(null, null, panel);
	}

	/**
	 * Create a new about box.
	 * 
	 * @param windowIcon The icon to use for the window.
	 * @param windowTitle The window title.
	 * @param panel The primary panel to display.
	 */
	public TKAboutWindow(String windowTitle, BufferedImage windowIcon, TKPanel panel) {
		super(windowTitle, windowIcon);
		setResizable(false);
		getContent().add(panel, TKCompassPosition.CENTER);
	}

	/** @return A newly created default about box content panel. */
	public static TKPanel getDefaultPanel() {
		return getDefaultPanel(TKAppWithUI.getSplashGraphic());
	}

	/**
	 * @param graphic The primary graphic to display.
	 * @return A newly created default about box content panel.
	 */
	public static TKPanel getDefaultPanel(BufferedImage graphic) {
		TKPanel content = new TKPanel(new TKColumnLayout());

		content.setBorder(new TKEmptyBorder(10));
		if (graphic != null) {
			content.add(new TKLabel(graphic));
		}
		content.add(new TKLabel(TKApp.getShortVersionBanner(), TKFont.CONTROL_FONT_KEY, TKAlignment.CENTER, true));
		content.add(new TKLabel(TKApp.getCopyrightBanner(true), TKAlignment.CENTER, true));
		return content;
	}

	/** Display the about box. */
	public void display() {
		Rectangle bounds = getGraphicsConfiguration().getBounds();
		Dimension size;

		pack();
		size = getSize();
		setLocation((bounds.width - size.width) / 2, (bounds.height - size.height) / 3);
		TKGraphics.forceOnScreen(this);
		setVisible(true);
	}
}
