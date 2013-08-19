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

import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;

import java.awt.AWTEvent;
import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.font.FontRenderContext;

/** A standard slider. */
public class TKSlider extends TKPanel implements TKSliderFormatter {
	private static final int	DISABLE_PERCENTAGE	= 50;
	private static final int	TIP_GAP				= 3;
	private static final int	GAP					= 1;
	private static final int	TICK_HEIGHT			= 8;
	private static final int	TRACK_HEIGHT		= 5;
	private static final int	KNOB_EXCESS			= 5;
	private static final int	KNOB_HEIGHT			= TRACK_HEIGHT + KNOB_EXCESS * 2;
	private static final int	KNOB_WIDTH			= 31;
	private static final int	NON_FONT_HEIGHT		= GAP + TICK_HEIGHT + TRACK_HEIGHT + KNOB_EXCESS;
	private TKSliderFormatter	mFormatter;
	private boolean				mContinuous;
	private int					mValue;
	private int[]				mTickMarks;
	private int[]				mTickMarkPositions;
	private boolean				mPressed;
	private boolean				mOver;
	private double				mMultiplier;
	private int					mFontAscent;

	/**
	 * Create a new slider.
	 * 
	 * @param tickMarks The tick mark values. At least 2 must be defined.
	 * @param formatter The value formatter for the slider.
	 * @param continuous Whether continous values are allowed (i.e. values other than those provided
	 *            by the tick marks).
	 */
	public TKSlider(int[] tickMarks, TKSliderFormatter formatter, boolean continuous) {
		super();

		if (tickMarks.length < 2) {
			throw new IllegalArgumentException("tickMarks must have at least two elements"); //$NON-NLS-1$
		}

		mFormatter = formatter == null ? this : formatter;
		mFontAscent = TKFont.getFontMetrics(TKFont.lookup(TKFont.TEXT_FONT_KEY)).getAscent();
		mContinuous = continuous;
		mValue = tickMarks[0];
		mTickMarks = new int[tickMarks.length];
		mTickMarkPositions = new int[tickMarks.length];
		System.arraycopy(tickMarks, 0, mTickMarks, 0, tickMarks.length);

		enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
	}

	private int getClosestIndex(int x) {
		int delta = Integer.MAX_VALUE;
		int which = 0;

		if (x > 0) {
			if (x > getWidth()) {
				which = mTickMarks.length - 1;
			} else {
				for (int i = 0; i < mTickMarks.length; i++) {
					int tmp = mTickMarkPositions[i] - x;

					if (tmp < 0) {
						tmp = -tmp;
					}
					if (tmp < delta) {
						delta = tmp;
						which = i;
					}
				}
			}
		}
		return which;
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = getPreferredSizeSelf();

		size.width = MAX_SIZE;
		return size;
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return getPreferredSizeSelf();
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		int halfKnobWidth = KNOB_WIDTH / 2;
		int tickZero = mTickMarks[0];
		int range = mTickMarks[mTickMarks.length - 1] - tickZero;
		int width = insets.left + insets.right + mTickMarks.length;
		int height = NON_FONT_HEIGHT + insets.top + insets.bottom + mFontAscent;
		int max = 0;

		for (int i = 0; i < mTickMarks.length - 1; i++) {
			Font font = TKFont.lookup(TKFont.TEXT_FONT_KEY);
			int leftSize = TKTextDrawing.getPreferredSize(font, null, mFormatter.getFormattedValue(this, mTickMarks[i], true)).width / 2;
			int rightSize = TKTextDrawing.getPreferredSize(font, null, mFormatter.getFormattedValue(this, mTickMarks[i + 1], true)).width / 2;
			int cur = (int) ((leftSize + rightSize) * range / (double) (mTickMarks[i + 1] - mTickMarks[i]));

			if (cur > max) {
				max = cur;
			}
			if (i == 0) {
				if (halfKnobWidth > leftSize) {
					width += halfKnobWidth;
				} else {
					width += leftSize;
				}
			}
			if (i == mTickMarks.length - 2) {
				if (halfKnobWidth > rightSize) {
					width += halfKnobWidth;
				} else {
					width += rightSize;
				}
			}
		}
		width += max;

		return new Dimension(width, height);
	}

	@Override public String getToolTipText(MouseEvent event) {
		return getToolTipText();
	}

	/** @return The current value. */
	public int getValue() {
		return mValue;
	}

	/** @param value The new current value. */
	public void setValue(int value) {
		if (mContinuous) {
			if (value < mTickMarks[0]) {
				value = mTickMarks[0];
			} else if (value > mTickMarks[mTickMarks.length - 1]) {
				value = mTickMarks[mTickMarks.length - 1];
			}
		} else {
			int closest = 0;
			int delta = Integer.MAX_VALUE;

			for (int element : mTickMarks) {
				int thisDelta = element - value;

				if (thisDelta < 0) {
					thisDelta = -thisDelta;
				}

				if (thisDelta < delta) {
					delta = thisDelta;
					closest = 0;
				}
			}

			value = mTickMarks[closest];
		}

		if (mValue != value) {
			mValue = value;
			repaint();
			notifyActionListeners();
		}
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		boolean enabled = isEnabled();
		Insets insets = getInsets();
		Font font = g2d.getFont();
		FontRenderContext frc = g2d.getFontRenderContext();
		int tickZero = mTickMarks[0];
		int range = mTickMarks[mTickMarks.length - 1] - tickZero;
		int halfKnob = KNOB_WIDTH / 2;
		int width = getWidth() - (1 + insets.left + insets.right + halfKnob * 2);
		int left = insets.left + halfKnob;
		int tickTop = insets.top + mFontAscent + GAP;
		int pos;
		int y;

		mMultiplier = (double) width / (double) range;
		width = (int) (mMultiplier * (mTickMarks[mTickMarks.length - 1] - tickZero));

		// Draw the track, labels & tick marks
		g2d.setColor(enabled ? TKColor.SCROLL_BAR_FILL : TKColor.lighter(TKColor.SCROLL_BAR_FILL, DISABLE_PERCENTAGE));
		g2d.fillRect(left, tickTop + TICK_HEIGHT, width + 1, TRACK_HEIGHT);
		g2d.setColor(enabled ? TKColor.SLIDER_SHADOW : TKColor.lighter(TKColor.SLIDER_SHADOW, DISABLE_PERCENTAGE));
		g2d.drawLine(left + 1, tickTop + TICK_HEIGHT + 1, left + 1, tickTop + TICK_HEIGHT + 1 + TRACK_HEIGHT - 2);
		g2d.drawLine(left + 1, tickTop + TICK_HEIGHT + 1, left + 1 + width - 2, tickTop + TICK_HEIGHT + 1);
		g2d.setColor(enabled ? TKColor.SCROLL_BAR_LINE : TKColor.lighter(TKColor.SCROLL_BAR_LINE, DISABLE_PERCENTAGE));
		g2d.drawRect(left, tickTop + TICK_HEIGHT, width, TRACK_HEIGHT - 1);
		for (int i = 0; i < mTickMarks.length; i++) {
			g2d.setColor(enabled ? TKColor.SCROLL_BAR_LINE : TKColor.lighter(TKColor.SCROLL_BAR_LINE, DISABLE_PERCENTAGE));
			String value = mFormatter.getFormattedValue(this, mTickMarks[i], true);

			pos = left + (int) (mMultiplier * (mTickMarks[i] - tickZero));
			g2d.drawLine(pos, tickTop, pos, tickTop + TICK_HEIGHT);
			mTickMarkPositions[i] = pos;
			g2d.setColor(enabled ? Color.black : TKColor.lighter(Color.black, DISABLE_PERCENTAGE));
			g2d.drawString(value, pos + 1 - TKTextDrawing.getWidth(font, frc, value) / 2, insets.top + mFontAscent);
		}

		// Draw the knob
		if (mPressed || mOver) {
			g2d.setColor(TKColor.CONTROL_ROLL);
		} else {
			g2d.setColor(enabled ? TKColor.CONTROL_FILL : TKColor.lighter(TKColor.CONTROL_FILL, DISABLE_PERCENTAGE));
		}

		pos = left + (int) (mMultiplier * (mValue - tickZero));
		y = tickTop + TICK_HEIGHT - KNOB_EXCESS;
		g2d.fillPolygon(new int[] { pos, pos + 3, pos + halfKnob - 2, pos + halfKnob, pos + halfKnob, pos + halfKnob - 2, pos + 2 - halfKnob, pos - halfKnob, pos - halfKnob, pos + 2 - halfKnob, pos - 3 }, new int[] { y - 3, y, y, y + 2, y + KNOB_HEIGHT - 3, y + KNOB_HEIGHT - 1, y + KNOB_HEIGHT - 1, y + KNOB_HEIGHT - 3, y + 2, y, y }, 11);

		g2d.setColor(enabled ? TKColor.CONTROL_HIGHLIGHT : TKColor.lighter(TKColor.CONTROL_HIGHLIGHT, DISABLE_PERCENTAGE));
		g2d.drawLine(pos, y - 2, pos - 3, y + 1);
		g2d.drawLine(pos - 3, y + 1, pos + 2 - halfKnob, y + 1);
		g2d.drawLine(pos + 1 - halfKnob, y + 2, pos + 1 - halfKnob, y + KNOB_HEIGHT - 3);
		g2d.drawLine(pos + 3, y + 1, pos + halfKnob - 2, y + 1);

		g2d.setColor(enabled ? TKColor.CONTROL_SHADOW : TKColor.lighter(TKColor.CONTROL_SHADOW, DISABLE_PERCENTAGE));
		g2d.drawLine(pos + halfKnob - 1, y + 2, pos + halfKnob - 1, y + KNOB_HEIGHT - 3);
		g2d.drawLine(pos + halfKnob - 2, y + KNOB_HEIGHT - 2, pos + 2 - halfKnob, y + KNOB_HEIGHT - 2);

		g2d.setColor(enabled ? TKColor.CONTROL_LINE : TKColor.lighter(TKColor.CONTROL_LINE, DISABLE_PERCENTAGE));
		g2d.drawPolygon(new int[] { pos, pos + 3, pos + halfKnob - 2, pos + halfKnob, pos + halfKnob, pos + halfKnob - 2, pos + 2 - halfKnob, pos - halfKnob, pos - halfKnob, pos + 2 - halfKnob, pos - 3 }, new int[] { y - 3, y, y, y + 2, y + KNOB_HEIGHT - 3, y + KNOB_HEIGHT - 1, y + KNOB_HEIGHT - 1, y + KNOB_HEIGHT - 3, y + 2, y, y }, 11);

		int midY = y + KNOB_HEIGHT / 2;
		drawSlash(g2d, pos - 5, midY + 1, 3);
		drawSlash(g2d, pos - 3, midY + 3, 6);
		drawSlash(g2d, pos + 2, midY + 2, 3);

		if (mPressed) {
			String curValue = mFormatter.getFormattedValue(this, mValue, false);
			Font menuKeyFont = TKFont.lookup(TKFont.MENU_KEY_FONT_KEY);
			Dimension size = TKTextDrawing.getPreferredSize(menuKeyFont, frc, curValue);
			FontMetrics fm = g2d.getFontMetrics(menuKeyFont);
			Rectangle bounds = new Rectangle(pos + halfKnob + TIP_GAP, y, size.width + 2, size.height + 2);

			if (bounds.x + bounds.width - left > width) {
				bounds.x = pos - (halfKnob + TIP_GAP + bounds.width);
			}
			g2d.setColor(TKColor.TOOLTIP_BACKGROUND);
			g2d.fill(bounds);
			g2d.setColor(Color.black);
			g2d.draw(bounds);
			g2d.setFont(menuKeyFont);
			g2d.drawString(curValue, bounds.x + 2, bounds.y + fm.getAscent() + 1);
		}
	}

	private void drawSlash(Graphics2D g2d, int x, int y, int height) {
		boolean enabled = isEnabled();
		g2d.setPaint(enabled ? TKColor.SCROLL_BAR_LINE : TKColor.lighter(TKColor.SCROLL_BAR_LINE, DISABLE_PERCENTAGE));
		g2d.drawLine(x, y, x + height, y - height);
		g2d.setPaint(enabled ? TKColor.CONTROL_HIGHLIGHT : TKColor.lighter(TKColor.CONTROL_HIGHLIGHT, DISABLE_PERCENTAGE));
		g2d.drawLine(x + 1, y, x + 1 + height, y - height);
	}

	@SuppressWarnings("fallthrough") @Override public void processMouseEventSelf(MouseEvent event) {
		switch (event.getID()) {
			case MouseEvent.MOUSE_PRESSED:
				mPressed = true;
				repaint();
				// Intentional fall-through!
			case MouseEvent.MOUSE_DRAGGED:
				int x = event.getX();
				Insets insets = getInsets();

				if (!mContinuous || event.getY() < insets.top + mFontAscent + GAP + TICK_HEIGHT) {
					setValue(mTickMarks[getClosestIndex(x)]);
				} else {
					setValue(mTickMarks[0] + (int) ((x - (insets.left + KNOB_WIDTH / 2)) / mMultiplier));
				}
				break;
			case MouseEvent.MOUSE_RELEASED:
				mPressed = false;
				repaint();
				break;
			case MouseEvent.MOUSE_ENTERED:
				mOver = true;
				repaint();
				break;
			case MouseEvent.MOUSE_EXITED:
				mOver = false;
				repaint();
				break;
		}
	}

	public String getFormattedValue(TKSlider slider, int value, boolean forTickMark) {
		return Integer.toString(value);
	}
}
