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

import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.utility.TKRectUtils;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.border.TKBevelBorder;

import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;

/** Provides a standard progress bar. */
public class TKProgressBar extends TKPanel implements Runnable {
	private static final int	MAX_STATE	= 10;
	private long				mCurrent;
	private long				mMaximum;
	private int					mLastPixelWidth;
	private long				mShowAfter;
	private boolean				mRunning;
	private boolean				mTerminate;
	private int					mState;

	/** Create a new indeterminate progress bar. */
	public TKProgressBar() {
		this(0);
	}

	/**
	 * Create a new progress bar.
	 * 
	 * @param length The length to use.
	 */
	public TKProgressBar(int length) {
		super();
		setOpaque(true);
		mMaximum = length < 0 ? 0 : length;
		mLastPixelWidth = -1;
	}

	/** @return The length of the task. */
	public int getLength() {
		return (int) mMaximum;
	}

	/**
	 * Sets the length of the task. This also resets any progress made back to the start.
	 * 
	 * @param length The length of the task.
	 */
	public void setLength(int length) {
		setLength(length, 0);
	}

	/**
	 * Sets the length of the task. This also resets any progress made back to the start.
	 * 
	 * @param length The length of the task.
	 * @param showAfter Only show the progress bar after this number of milliseconds have elapsed.
	 */
	public void setLength(int length, long showAfter) {
		if (length < 0) {
			length = 0;
		}
		if (length != mMaximum) {
			mCurrent = 0;
			mMaximum = length;
			mShowAfter = System.currentTimeMillis() + showAfter - 1;
			paintImmediately();
		}
	}

	/** Turns this progress bar off. */
	public void finished() {
		stop();
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		Insets borderInsets = TKBevelBorder.getSharedBorder(false).getBorderInsets(this);

		return new Dimension(100 + insets.left + insets.right + borderInsets.left + borderInsets.right, 7 + insets.top + insets.bottom + borderInsets.top + borderInsets.bottom);
	}

	/** @return The current value of the progress bar. */
	public int getValue() {
		return (int) mCurrent;
	}

	/** Increments the progress bar by one unit. */
	public void increment() {
		increment(1);
	}

	/**
	 * Increments the progress bar by the specified amount.
	 * 
	 * @param amount The amount to increment the progress bar by.
	 */
	public void increment(int amount) {
		setValue(getValue() + amount);
	}

	@Override protected synchronized void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		if (mMaximum < 1 && mRunning && !mTerminate || mMaximum > 0 && mShowAfter < System.currentTimeMillis()) {
			Rectangle bounds = getLocalInsetBounds();
			TKBevelBorder border = TKBevelBorder.getSharedBorder(false);
			Insets borderInsets = border.getBorderInsets(this);

			border.paintBorder(this, g2d, bounds.x, bounds.y, bounds.width, bounds.height);
			bounds.x += borderInsets.left;
			bounds.y += borderInsets.top;
			bounds.width -= borderInsets.left + borderInsets.right;
			bounds.height -= borderInsets.top + borderInsets.bottom;

			g2d.setClip(TKRectUtils.intersection(bounds, g2d.getClipBounds()));
			g2d.setColor(TKColor.PROGRESS_BAR);

			if (mMaximum > 0) {
				if (mCurrent > 0) {
					bounds.width = (int) (bounds.width * mCurrent / mMaximum);
					g2d.fill(bounds);
				}
			} else {
				int state = mState;

				if (state < 0) {
					state = -state;
				}
				bounds.width /= MAX_STATE;
				bounds.x += state * bounds.width;
				g2d.fill(bounds);
			}
		}
	}

	/** @param value The new value of this progress bar. */
	public void setValue(int value) {
		if (value < 0) {
			value = 0;
		}
		if (value > mMaximum) {
			value = (int) mMaximum;
		}
		if (value != mCurrent) {
			Rectangle bounds = getLocalInsetBounds();
			int newPixelWidth;

			mCurrent = value;
			newPixelWidth = (int) (bounds.width * mCurrent / mMaximum);
			if (newPixelWidth != mLastPixelWidth && mShowAfter < System.currentTimeMillis()) {
				paintImmediately();
				mLastPixelWidth = newPixelWidth;
			}
		}
	}

	/** Turn an indeterminate progress bar on. */
	public void start() {
		start(0);
	}

	/**
	 * Turn an indeterminate progress bar on.
	 * 
	 * @param showAfter Only show the progress bar after this number of milliseconds have elapsed.
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
		setLength(0);
		if (mRunning) {
			mTerminate = true;
			paintImmediately();
		}
	}

	public synchronized void run() {
		if (mRunning && !mTerminate) {
			if (++mState >= MAX_STATE) {
				mState = 2 - MAX_STATE;
			}
			paintImmediately();
			TKTimerTask.schedule(this, 100, false);
		} else {
			mRunning = mTerminate = false;
		}
	}
}
