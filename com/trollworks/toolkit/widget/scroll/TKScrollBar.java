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

package com.trollworks.toolkit.widget.scroll;

import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.widget.TKPanel;

import java.awt.AWTEvent;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;

/** A standard scroll bar. */
public class TKScrollBar extends TKPanel implements Runnable {
	private static final int	BAR_SIZE			= 14;	// Should be even.
	private static final int	MINIMUM_THUMB_SIZE	= 10;
	private static final int	SLASH_HEIGHT		= 5;
	private static final int	SNAPBACK_THRESHOLD	= 128;
	private static final int	NOTHING				= 0;
	private static final int	THUMB				= 1;
	private static final int	UP_LEFT				= 2;
	private static final int	DOWN_RIGHT			= 3;
	private static final int	UP_LEFT_PAGE		= 4;
	private static final int	DOWN_RIGHT_PAGE		= 5;
	private int					mMinimum;
	private int					mMaximum;
	private int					mCurrent;
	private boolean				mVertical;
	private int					mBlockScrollIncrement;
	private int					mUnitScrollIncrement;
	private int					mContentSize;
	private int					mContentViewSize;
	private TKScrollBarOwner	mOwner;
	private int					mHighlight;
	private boolean				mTimerPending;
	private boolean				mFirstTimer;
	private int					mHit;
	private int					mRollHit;
	private int					mOriginalValue;
	private Rectangle			mHitBounds;
	private int					mThumbOffset;

	/**
	 * Creates a scroll bar.
	 * 
	 * @param vertical Pass in <code>true</code> for a vertical scroll bar and <code>false</code>
	 *            for a horizontal one.
	 * @param min The minimum value for the scroll bar.
	 * @param max The maximum value for the scroll bar.
	 */
	public TKScrollBar(boolean vertical, int min, int max) {
		this(null, vertical, min, max, min);
	}

	/**
	 * Creates a scroll bar.
	 * 
	 * @param vertical Pass in <code>true</code> for a vertical scroll bar and <code>false</code>
	 *            for a horizontal one.
	 * @param min The minimum value for the scroll bar.
	 * @param max The maximum value for the scroll bar.
	 * @param initial The initial value for the scroll bar.
	 */
	public TKScrollBar(boolean vertical, int min, int max, int initial) {
		this(null, vertical, min, max, initial);
	}

	/**
	 * Creates a scroll bar.
	 * 
	 * @param owner The owning {@link TKScrollBarOwner}.
	 * @param vertical Pass in <code>true</code> for a vertical scroll bar and <code>false</code>
	 *            for a horizontal one.
	 * @param min The minimum value for the scroll bar.
	 * @param max The maximum value for the scroll bar.
	 * @param initial The initial value for the scroll bar.
	 */
	public TKScrollBar(TKScrollBarOwner owner, boolean vertical, int min, int max, int initial) {
		super();

		if (max < min) {
			max = min;
		}
		if (initial < min) {
			initial = min;
		} else if (initial > max) {
			initial = max;
		}
		mOwner = owner;
		mMinimum = min;
		mMaximum = max;
		mCurrent = initial;
		mBlockScrollIncrement = 1;
		mUnitScrollIncrement = 1;
		mVertical = vertical;
		mHighlight = NOTHING;
		mHit = NOTHING;
		setOpaque(true);
		setCursor(Cursor.getDefaultCursor());
		enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
	}

	/**
	 * @param upLeftDirection Whether the scroll is up/left or down/right.
	 * @return The block increment.
	 */
	public int getBlockScrollIncrement(boolean upLeftDirection) {
		if (mOwner != null) {
			return mOwner.getBlockScrollIncrement(mVertical, upLeftDirection);
		}
		return upLeftDirection ? -mBlockScrollIncrement : mBlockScrollIncrement;
	}

	/** @return The content size. */
	public int getContentSize() {
		if (mOwner != null) {
			return mOwner.getContentSize(mVertical);
		}
		return mContentSize < 1 ? 1 + mMaximum - mMinimum : mContentSize;
	}

	/** @return The content view port size. */
	public int getContentViewSize() {
		if (mOwner != null) {
			return mOwner.getContentViewSize(mVertical);
		}
		return mContentViewSize < 1 ? 1 : mContentViewSize;
	}

	private Rectangle getControlArrowBounds(boolean upLeftDirection) {
		Rectangle bounds = getLocalInsetBounds();

		if (upLeftDirection) {
			return new Rectangle(bounds.x, bounds.y, BAR_SIZE, BAR_SIZE);
		} else if (mVertical) {
			return new Rectangle(bounds.x, bounds.y + bounds.height - (BAR_SIZE + 1), BAR_SIZE, BAR_SIZE);
		}
		return new Rectangle(bounds.x + bounds.width - (BAR_SIZE + 1), bounds.y, BAR_SIZE, BAR_SIZE);
	}

	/** @return The current value. */
	public int getCurrentValue() {
		return mCurrent;
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = getPreferredSizeSelf();

		if (mVertical) {
			size.height = MAX_SIZE;
		} else {
			size.width = MAX_SIZE;
		}
		return size;
	}

	/** @return The maximum value. */
	public int getMaximumValue() {
		return mMaximum;
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return getPreferredSizeSelf();
	}

	/** @return The minimum value. */
	public int getMinimumValue() {
		return mMinimum;
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		Dimension size = new Dimension(insets.left + insets.right, insets.top + insets.bottom);

		if (mVertical) {
			size.width += BAR_SIZE + 1;
			size.height += BAR_SIZE;
		} else {
			size.width += BAR_SIZE;
			size.height += BAR_SIZE + 1;
		}
		return size;
	}

	/**
	 * Maps the specified location in the content to a position in the gutter.
	 * 
	 * @param location The location to map.
	 * @return The mapped position.
	 */
	public int mapContentLocationToGutter(int location) {
		Rectangle bounds = getLocalInsetBounds();
		long gutterSize = (mVertical ? bounds.height : bounds.width) - BAR_SIZE * 2;
		long contentSize = getContentSize();

		return BAR_SIZE + (int) (location * gutterSize / contentSize);
	}

	/** @return The bounds of the thumb in the scrollbar's coordinate system. */
	public Rectangle getThumbBounds() {
		if (mMinimum != mMaximum) {
			int thumbSize = getContentViewSize();
			int contentSize = getContentSize();

			if (thumbSize < contentSize) {
				Rectangle bounds = getLocalInsetBounds();
				int gutterSize = (mVertical ? bounds.height : bounds.width) - BAR_SIZE * 2;

				thumbSize = gutterSize * thumbSize / contentSize;
				if (thumbSize < MINIMUM_THUMB_SIZE) {
					thumbSize = MINIMUM_THUMB_SIZE;
				}
				if (thumbSize < gutterSize) {
					int range = 1 + mMaximum - mMinimum;
					int x = bounds.x;
					int y = bounds.y;
					int width = mVertical ? bounds.width - 1 : thumbSize + 1;
					int height = mVertical ? thumbSize + 1 : bounds.height - 1;

					range = BAR_SIZE + (gutterSize - thumbSize) * mCurrent / range;
					if (mVertical) {
						y += range;
					} else {
						x += range;
					}
					return new Rectangle(x, y, width, height);
				}
			}
		}

		return new Rectangle();
	}

	/**
	 * @param upLeftDirection Whether the scroll is up/left or down/right.
	 * @return The unit increment.
	 */
	public int getUnitScrollIncrement(boolean upLeftDirection) {
		if (mOwner != null) {
			return mOwner.getUnitScrollIncrement(mVertical, upLeftDirection);
		}
		return upLeftDirection ? -mUnitScrollIncrement : mUnitScrollIncrement;
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The value the scroll bar would be set to if the thumb was moved to the specified
	 *         location (without accounting for the height of the thumb).
	 */
	public int getValueAtLocation(int x, int y) {
		if (mMinimum != mMaximum) {
			Rectangle bounds = getLocalInsetBounds();
			int pos = mVertical ? y : x;
			int min = (mVertical ? bounds.y : bounds.x) + BAR_SIZE;
			int max = (mVertical ? bounds.y + bounds.height : bounds.x + bounds.width) - BAR_SIZE * 2;

			if (pos <= min) {
				return mMinimum;
			}
			if (pos >= max) {
				return mMaximum;
			}

			return mMinimum + (pos - min) * (mMaximum - mMinimum) / (max - min);
		}

		return mMinimum;
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The value the scroll bar would be set to if the thumb was moved to the specified
	 *         location, accounting for the thumb height.
	 */
	public int getValueAtLocationAdjusted(int x, int y) {
		if (mMinimum != mMaximum) {
			Rectangle bounds = getLocalInsetBounds();
			int pos = mVertical ? y : x;
			int min = (mVertical ? bounds.y : bounds.x) + BAR_SIZE;
			int max = mVertical ? bounds.y + bounds.height - (BAR_SIZE + getThumbBounds().height) : bounds.x + bounds.width - (BAR_SIZE + getThumbBounds().width);
			int delta = max - min;

			return mMinimum + (delta > 0 ? (pos - min) * (mMaximum - mMinimum) / (max - min) : 0);
		}

		return mMinimum;
	}

	/** @return <code>true</code> if this scroll bar is vertically oriented. */
	public boolean isVertical() {
		return mVertical;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Rectangle bounds = getLocalInsetBounds();
		Rectangle thumbBounds = getThumbBounds();

		// draw the bar in between scroll arrows
		g2d.setPaint(TKColor.SCROLL_BAR_FILL);
		g2d.fillRect(bounds.x, bounds.y, bounds.width - 1, bounds.height - 1);
		g2d.setPaint(TKColor.SCROLL_BAR_LINE);
		g2d.drawRect(bounds.x, bounds.y, bounds.width - 1, bounds.height - 1);

		// Draw the scroll thumb
		drawBeveledButton(g2d, thumbBounds, false, mRollHit == THUMB || mHit == THUMB);

		// Draw the thumb slashes
		int sX = thumbBounds.x + thumbBounds.width / 2 - SLASH_HEIGHT / 2;
		int sY = thumbBounds.y + thumbBounds.height / 2 + SLASH_HEIGHT / 2;

		if (mVertical) {
			if (thumbBounds.height > SLASH_HEIGHT * 3 + 4) {
				drawSlash(g2d, sX - 1, sY, true);
				drawSlash(g2d, sX - 1, sY + SLASH_HEIGHT, true);
				drawSlash(g2d, sX - 1, sY - SLASH_HEIGHT, true);
			}
		} else {
			if (thumbBounds.width > SLASH_HEIGHT * 3 + 4) {
				drawSlash(g2d, sX, sY + 1, false);
				drawSlash(g2d, sX + SLASH_HEIGHT, sY + 1, false);
				drawSlash(g2d, sX - SLASH_HEIGHT, sY + 1, false);
			}
		}

		// Draw the arrow buttons
		Rectangle buttonBounds = getControlArrowBounds(true);
		drawBeveledButton(g2d, buttonBounds, mHit == UP_LEFT, mRollHit == UP_LEFT);
		g2d.setPaint(TKColor.CONTROL_ICON);
		drawControlArrow(g2d, buttonBounds.x, buttonBounds.y, true);
		buttonBounds = getControlArrowBounds(false);
		drawBeveledButton(g2d, buttonBounds, mHit == DOWN_RIGHT, mRollHit == DOWN_RIGHT);
		g2d.setPaint(TKColor.CONTROL_ICON);
		drawControlArrow(g2d, buttonBounds.x, buttonBounds.y, false);
	}

	private void drawBeveledButton(Graphics2D g2d, Rectangle bounds, boolean pressed, boolean rollHighlight) {
		g2d.setPaint(pressed ? TKColor.CONTROL_PRESSED_FILL : rollHighlight ? TKColor.CONTROL_ROLL : TKColor.CONTROL_FILL);
		g2d.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
		g2d.setPaint(TKColor.CONTROL_LINE);
		g2d.drawRect(bounds.x, bounds.y, bounds.width, bounds.height);

		if (pressed) {
			g2d.setPaint(TKColor.CONTROL_HIGHLIGHT);
		} else {
			g2d.setPaint(TKColor.CONTROL_SHADOW);
		}

		int x = bounds.x + bounds.width - 1;
		int y = bounds.y + bounds.height - 1;
		g2d.drawLine(x, bounds.y + 1, x, y);
		g2d.drawLine(bounds.x + 1, y, x, y);

		if (!pressed) {
			g2d.setPaint(TKColor.CONTROL_HIGHLIGHT);
			g2d.drawLine(bounds.x + 1, bounds.y + 1, bounds.x + 1, y);
			g2d.drawLine(bounds.x + 1, bounds.y + 1, x, bounds.y + 1);
		}
	}

	private void drawSlash(Graphics2D g2d, int x, int y, boolean vertical) {
		g2d.setPaint(TKColor.SCROLL_BAR_LINE);
		g2d.drawLine(x, y, x + SLASH_HEIGHT, y - SLASH_HEIGHT);
		g2d.setPaint(TKColor.CONTROL_HIGHLIGHT);
		if (vertical) {
			g2d.drawLine(x, y + 1, x + SLASH_HEIGHT, y + 1 - SLASH_HEIGHT);
		} else {
			g2d.drawLine(x + 1, y, x + 1 + SLASH_HEIGHT, y - SLASH_HEIGHT);
		}
	}

	private void drawControlArrow(Graphics2D g2d, int x, int y, boolean upLeftDirection) {
		int midPoint;
		int pos;
		int i;
		int adj = upLeftDirection ? -2 : 2;

		if (mVertical) {
			midPoint = x + BAR_SIZE / 2;
			pos = y + BAR_SIZE / 2 + adj;
			for (i = 0; i < 4; i++) {
				if (upLeftDirection) {
					g2d.drawLine(midPoint - i, pos, midPoint + i, pos);
					pos++;
				} else {
					g2d.drawLine(midPoint - i, pos, midPoint + i, pos);
					pos--;
				}
			}
			g2d.drawLine(midPoint - 2, pos, midPoint - 2, pos);
			g2d.drawLine(midPoint + 2, pos, midPoint + 2, pos);
		} else {
			midPoint = y + BAR_SIZE / 2;
			pos = x + BAR_SIZE / 2 + adj;
			for (i = 0; i < 4; i++) {
				if (upLeftDirection) {
					g2d.drawLine(pos, midPoint - i, pos, midPoint + i);
					pos++;
				} else {
					g2d.drawLine(pos, midPoint - i, pos, midPoint + i);
					pos--;
				}
			}
			g2d.drawLine(pos, midPoint - 2, pos, midPoint - 2);
			g2d.drawLine(pos, midPoint + 2, pos, midPoint + 2);
		}
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		int x = event.getX();
		int y = event.getY();
		Rectangle thumbBounds = getThumbBounds();

		switch (event.getID()) {
			case MouseEvent.MOUSE_PRESSED:
				if (thumbBounds.contains(x, y)) {
					mOriginalValue = getCurrentValue();
					mHit = THUMB;
					mHighlight = mHit;
					mThumbOffset = mVertical ? y - thumbBounds.y : x - thumbBounds.x;
				} else {
					mHitBounds = getControlArrowBounds(true);

					if (mHitBounds.contains(x, y)) {
						mHit = UP_LEFT;
					} else {
						mHitBounds = getControlArrowBounds(false);

						if (mHitBounds.contains(x, y)) {
							mHit = DOWN_RIGHT;
						} else if (mVertical && y < thumbBounds.y || !mVertical && x < thumbBounds.x) {
							mHit = UP_LEFT_PAGE;
						} else {
							mHit = DOWN_RIGHT_PAGE;
						}
					}

					mHighlight = mHit;
					mFirstTimer = true;
					runScroll();
				}
				repaint();
				break;
			case MouseEvent.MOUSE_MOVED:
				int oldRollHit = mRollHit;
				mRollHit = NOTHING;

				if (thumbBounds.contains(x, y)) {
					mRollHit = THUMB;
				} else {
					if (getControlArrowBounds(true).contains(x, y)) {
						mRollHit = UP_LEFT;
					} else if (getControlArrowBounds(false).contains(x, y)) {
						mRollHit = DOWN_RIGHT;
					}
				}

				if (mRollHit != oldRollHit) {
					repaint();
				}
				break;
			case MouseEvent.MOUSE_EXITED:
				if (mRollHit != NOTHING) {
					mRollHit = NOTHING;
					repaint();
				}
				break;
			case MouseEvent.MOUSE_RELEASED:
				mHighlight = NOTHING;
				mHit = NOTHING;
				repaint();
				break;
			case MouseEvent.MOUSE_DRAGGED:
				switch (mHit) {
					case THUMB:
						if (mMinimum != mMaximum) {
							int tmp = mVertical ? x : y;

							if (tmp >= -SNAPBACK_THRESHOLD && tmp < (mVertical ? getWidth() : getHeight()) + SNAPBACK_THRESHOLD) {
								mHighlight = mHit;
								if (mVertical) {
									setCurrentValue(getValueAtLocationAdjusted(x, y - mThumbOffset));
								} else {
									setCurrentValue(getValueAtLocationAdjusted(x - mThumbOffset, y));
								}
							} else {
								mHighlight = NOTHING;
								setCurrentValue(mOriginalValue);
							}
						}
						break;
					case UP_LEFT:
					case DOWN_RIGHT:
						boolean inside = mHitBounds.contains(event.getPoint());

						if (mHighlight == mHit && !inside) {
							mHighlight = NOTHING;
							repaint();
						} else if (mHighlight != mHit && inside) {
							mHighlight = mHit;
							runScroll();
						}
						break;
				}
		}
	}

	public void run() {
		mTimerPending = false;
		mFirstTimer = false;
		runScroll();
	}

	private void runScroll() {
		if (mHighlight == mHit) {
			boolean repeat = true;

			switch (mHit) {
				case UP_LEFT:
					scroll(true, false);
					break;
				case DOWN_RIGHT:
					scroll(false, false);
					break;
				case UP_LEFT_PAGE:
					scroll(true, true);
					break;
				case DOWN_RIGHT_PAGE:
					scroll(false, true);
					break;
				default:
					repeat = false;
					break;
			}
			if (repeat && !mTimerPending) {
				mTimerPending = true;
				TKTimerTask.schedule(this, TKScrollPanel.AUTOSCROLL_DELAY * (mFirstTimer ? 4 : 1));
			}
		}
	}

	/**
	 * Scroll.
	 * 
	 * @param upLeftDirection Pass in <code>true</code> to scroll in the up or left direction,
	 *            <code>false</code> for the down or right direction.
	 * @param page Pass in <code>true</code> to scroll a full block at a time, <code>false</code>
	 *            to scroll one unit.
	 */
	public void scroll(boolean upLeftDirection, boolean page) {
		setCurrentValue(getCurrentValue() + (page ? getBlockScrollIncrement(upLeftDirection) : getUnitScrollIncrement(upLeftDirection)));
	}

	/**
	 * @param value The block increment. This will be ignored if an owner has been set.
	 */
	public void setBlockScrollIncrement(int value) {
		mBlockScrollIncrement = value;
	}

	/**
	 * @param value The content size. This will be ignored if an owner has been set.
	 */
	public void setContentSize(int value) {
		if (value != mContentSize) {
			mContentSize = value;
			repaint();
		}
	}

	/**
	 * @param value The content view port size. This will be ignored if an owner has been set.
	 */
	public void setContentViewSize(int value) {
		if (value != mContentViewSize) {
			mContentViewSize = value;
			repaint();
		}
	}

	/** @param value The current value. */
	public void setCurrentValue(int value) {
		if (value < mMinimum) {
			value = mMinimum;
		} else if (value > mMaximum) {
			value = mMaximum;
		}
		if (value != mCurrent) {
			mCurrent = value;
			repaint();
			notifyActionListeners();
		}
	}

	/** @param value The maximum value. */
	public void setMaximumValue(int value) {
		if (value != mMaximum) {
			mMaximum = value;
			if (mMinimum > mMaximum) {
				mMinimum = mMaximum;
			}
			repaint();
			if (mCurrent > mMaximum) {
				setCurrentValue(mMaximum);
			}
		}
	}

	/** @param value The minimum value. */
	public void setMinimumValue(int value) {
		if (value != mMinimum) {
			mMinimum = value;
			if (mMaximum < mMinimum) {
				mMaximum = mMinimum;
			}
			repaint();
			if (mCurrent < mMinimum) {
				setCurrentValue(mMinimum);
			}
		}
	}

	/**
	 * @param value The unit increment. This will be ignored if an owner has been set.
	 */
	public void setUnitScrollIncrement(int value) {
		mUnitScrollIncrement = value;
	}
}
