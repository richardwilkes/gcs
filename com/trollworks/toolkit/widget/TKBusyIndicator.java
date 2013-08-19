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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.utility.TKTimerTask;

import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.image.BufferedImage;

/** Provides a standard busy indicator. */
public class TKBusyIndicator extends TKPanel implements Runnable {
	private int		mState;
	private boolean	mRunning;
	private boolean	mTerminate;

	/** Create a new busy indicator. */
	public TKBusyIndicator() {
		super();
		setOpaque(true);
		mState = 1;
	}

	/** Turn this busy indicator on. */
	public void start() {
		start(0);
	}

	/**
	 * Turn this busy indicator on.
	 * 
	 * @param showAfter Only show the busy indicator after this number of milliseconds have elapsed.
	 */
	public synchronized void start(long showAfter) {
		if (!mRunning) {
			mRunning = true;
			if (mTerminate) {
				mTerminate = false;
				paintImmediately();
			} else {
				TKTimerTask.schedule(this, showAfter, false);
			}
		} else {
			mTerminate = false;
		}
	}

	/** Turn this busy indicator off. */
	public synchronized void stop() {
		if (mRunning) {
			mTerminate = true;
			paintImmediately();
		}
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		BufferedImage icon = TKImage.getBusyIcon(0);

		return new Dimension(icon.getWidth() + insets.left + insets.right, icon.getHeight() + insets.top + insets.bottom);
	}

	@Override protected synchronized void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		if (mRunning && !mTerminate) {
			BufferedImage icon = TKImage.getBusyIcon(mState);
			Insets insets = getInsets();

			g2d.drawImage(icon, insets.left, insets.right, null);
		}
	}

	public synchronized void run() {
		if (mRunning && !mTerminate) {
			mState = ++mState % 8;
			paintImmediately();
			TKTimerTask.schedule(this, 100, false);
		} else {
			mRunning = mTerminate = false;
		}
	}
}
