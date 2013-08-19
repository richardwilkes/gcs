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

import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.utility.TKPlatform;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;

import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Rectangle;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;

/** A standard splash screen. */
public class TKSplashWindow extends TKWindow implements Runnable, TKUserInputMonitor {
	private static final long	WAIT_TIME	= 5000;
	private Runnable			mTask;

	/** Create a new splash screen. */
	public TKSplashWindow() {
		this(null, TKAboutWindow.getDefaultPanel());
	}

	/**
	 * Create a new splash screen.
	 * 
	 * @param windowTitle The window title.
	 * @param graphic The primary graphic to display.
	 */
	public TKSplashWindow(String windowTitle, BufferedImage graphic) {
		this(windowTitle, new TKLabel(graphic));
	}

	/**
	 * Create a new splash screen.
	 * 
	 * @param panel The primary panel to display.
	 */
	public TKSplashWindow(TKPanel panel) {
		this(null, panel);
	}

	/**
	 * Create a new splash screen.
	 * 
	 * @param windowTitle The window title.
	 * @param panel The primary panel to display.
	 */
	public TKSplashWindow(String windowTitle, TKPanel panel) {
		super(windowTitle != null ? windowTitle : TKApp.getName(), null);
		setUndecorated(true);

		TKPanel content = getContent();
		if (!TKPlatform.isMacintosh()) {
			content.setBorder(TKLineBorder.getSharedBorder(true));
		}
		content.setLayout(new TKCompassLayout());
		content.add(panel, TKCompassPosition.CENTER);
		TKUserInputManager.addMonitor(this);
	}

	/**
	 * Display the splash screen and start the timer.
	 * 
	 * @param task A task to execute when the splash screen goes away.
	 */
	public void display(Runnable task) {
		Rectangle bounds = getGraphicsConfiguration().getBounds();
		Dimension size;

		pack();
		size = getSize();
		setLocation((bounds.width - size.width) / 2, (bounds.height - size.height) / 3);
		TKGraphics.forceOnScreen(this);
		setVisible(true);
		mTask = task;
		TKTimerTask.schedule(this, WAIT_TIME);
	}

	public void run() {
		TKUserInputManager.removeMonitor(this);
		if (!isClosed()) {
			dispose();
			if (mTask != null) {
				EventQueue.invokeLater(mTask);
			}
		}
	}

	public void userInputEventOccurred(InputEvent event) {
		int id = event.getID();

		if (id == MouseEvent.MOUSE_PRESSED || id == KeyEvent.KEY_PRESSED) {
			EventQueue.invokeLater(this);
		}
	}
}
