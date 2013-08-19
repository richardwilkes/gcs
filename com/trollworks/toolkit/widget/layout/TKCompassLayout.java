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

package com.trollworks.toolkit.widget.layout;

import com.trollworks.toolkit.collections.TKEnumExtractor;
import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;

/**
 * This is a replacement for <code>java.awt.BorderLayout</code> which honors various features of
 * {@link TKPanel}.
 */
public class TKCompassLayout implements LayoutManager2 {
	private int			mHGap;
	private int			mVGap;
	private Component	mNorth;
	private Component	mWest;
	private Component	mEast;
	private Component	mSouth;
	private Component	mCenter;

	/** Constructs a new {@link TKCompassLayout} with no gaps between components. */
	public TKCompassLayout() {
		this(0, 0);
	}

	/**
	 * Constructs a new {@link TKCompassLayout} with the specified gaps between components.
	 * 
	 * @param hgap The horizontal gap.
	 * @param vgap The vertical gap.
	 */
	public TKCompassLayout(int hgap, int vgap) {
		mHGap = hgap;
		mVGap = vgap;
	}

	public void addLayoutComponent(Component comp, Object constraints) {
		synchronized (comp.getTreeLock()) {
			if (constraints == null) {
				mCenter = comp;
			} else if (constraints instanceof TKCompassPosition) {
				switch ((TKCompassPosition) constraints) {
					case NORTH:
						mNorth = comp;
						break;
					case EAST:
						mEast = comp;
						break;
					case SOUTH:
						mSouth = comp;
						break;
					case WEST:
						mWest = comp;
						break;
					case CENTER:
						mCenter = comp;
						break;
				}
			} else {
				throw new IllegalArgumentException("Invalid constraint: " + constraints); //$NON-NLS-1$
			}
		}
	}

	public void addLayoutComponent(String name, Component comp) {
		addLayoutComponent(comp, TKEnumExtractor.extract(name, TKCompassPosition.values()));
	}

	/** @return The horizontal gap between components. */
	public int getHGap() {
		return mHGap;
	}

	public float getLayoutAlignmentX(Container parent) {
		return Component.CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container parent) {
		return Component.CENTER_ALIGNMENT;
	}

	/** @return The vertical gap between components. */
	public int getVGap() {
		return mVGap;
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public void layoutContainer(Container target) {
		synchronized (target.getTreeLock()) {
			Insets insets = target.getInsets();
			int top = insets.top;
			int bottom = target.getHeight() - insets.bottom;
			int left = insets.left;
			int right = target.getWidth() - insets.right;
			int tmp;

			if (mNorth != null) {
				tmp = mNorth.getPreferredSize().height;
				mNorth.setBounds(left, top, right - left, tmp);
				top += tmp + mVGap;
			}

			if (mSouth != null) {
				tmp = mSouth.getPreferredSize().height;
				mSouth.setBounds(left, bottom - tmp, right - left, tmp);
				bottom -= tmp + mVGap;
			}

			if (mEast != null) {
				tmp = mEast.getPreferredSize().width;
				mEast.setBounds(right - tmp, top, tmp, bottom - top);
				right -= tmp + mHGap;
			}

			if (mWest != null) {
				tmp = mWest.getPreferredSize().width;
				mWest.setBounds(left, top, tmp, bottom - top);
				left += tmp + mHGap;
			}

			if (mCenter != null) {
				mCenter.setBounds(left, top, right - left, bottom - top);
			}
		}
	}

	public Dimension maximumLayoutSize(Container target) {
		synchronized (target.getTreeLock()) {
			Insets insets = target.getInsets();
			Dimension mSize = mCenter == null ? new Dimension(0, 0) : new Dimension(mCenter.getMaximumSize());
			int tmp;

			TKPanel.sanitizeSize(mSize);
			if (mEast != null) {
				mSize.width += mEast.getPreferredSize().width + mHGap;
				tmp = mEast.getMaximumSize().height;
				if (tmp > mSize.height) {
					mSize.height = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			if (mWest != null) {
				mSize.width += mWest.getPreferredSize().width + mHGap;
				tmp = mWest.getMaximumSize().height;
				if (tmp > mSize.height) {
					mSize.height = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			if (mNorth != null) {
				mSize.height += mNorth.getPreferredSize().height + mVGap;
				tmp = mNorth.getMaximumSize().width;
				if (tmp > mSize.width) {
					mSize.width = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			if (mSouth != null) {
				mSize.height += mSouth.getPreferredSize().height + mVGap;
				tmp = mSouth.getMaximumSize().width;
				if (tmp > mSize.width) {
					mSize.width = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			mSize.width += insets.left + insets.right;
			mSize.height += insets.top + insets.bottom;

			return TKPanel.sanitizeSize(mSize);
		}
	}

	public Dimension minimumLayoutSize(Container target) {
		synchronized (target.getTreeLock()) {
			Insets insets = target.getInsets();
			Dimension mSize = mCenter == null ? new Dimension(0, 0) : new Dimension(mCenter.getMinimumSize());
			int tmp;

			TKPanel.sanitizeSize(mSize);
			if (mEast != null) {
				mSize.width += mEast.getPreferredSize().width + mHGap;
				tmp = mEast.getMinimumSize().height;
				if (tmp > mSize.height) {
					mSize.height = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			if (mWest != null) {
				mSize.width += mWest.getPreferredSize().width + mHGap;
				tmp = mWest.getMinimumSize().height;
				if (tmp > mSize.height) {
					mSize.height = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			if (mNorth != null) {
				mSize.height += mNorth.getPreferredSize().height + mVGap;
				tmp = mNorth.getMinimumSize().width;
				if (tmp > mSize.width) {
					mSize.width = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			if (mSouth != null) {
				mSize.height += mSouth.getPreferredSize().height + mVGap;
				tmp = mSouth.getMinimumSize().width;
				if (tmp > mSize.width) {
					mSize.width = tmp;
				}
				TKPanel.sanitizeSize(mSize);
			}

			mSize.width += insets.left + insets.right;
			mSize.height += insets.top + insets.bottom;

			return TKPanel.sanitizeSize(mSize);
		}
	}

	public Dimension preferredLayoutSize(Container target) {
		synchronized (target.getTreeLock()) {
			Insets insets = target.getInsets();
			Dimension pSize = mCenter == null ? new Dimension(0, 0) : new Dimension(mCenter.getPreferredSize());
			Dimension oSize;

			TKPanel.sanitizeSize(pSize);
			if (mEast != null) {
				oSize = mEast.getPreferredSize();
				pSize.width += oSize.width + mHGap;
				if (oSize.height > pSize.height) {
					pSize.height = oSize.height;
				}
				TKPanel.sanitizeSize(pSize);
			}

			if (mWest != null) {
				oSize = mWest.getPreferredSize();
				pSize.width += oSize.width + mHGap;
				if (oSize.height > pSize.height) {
					pSize.height = oSize.height;
				}
				TKPanel.sanitizeSize(pSize);
			}

			if (mNorth != null) {
				oSize = mNorth.getPreferredSize();
				pSize.height += oSize.height + mVGap;
				if (oSize.width > pSize.width) {
					pSize.width = oSize.width;
				}
				TKPanel.sanitizeSize(pSize);
			}

			if (mSouth != null) {
				oSize = mSouth.getPreferredSize();
				pSize.height += oSize.height + mVGap;
				if (oSize.width > pSize.width) {
					pSize.width = oSize.width;
				}
				TKPanel.sanitizeSize(pSize);
			}

			pSize.width += insets.left + insets.right;
			pSize.height += insets.top + insets.bottom;

			return TKPanel.sanitizeSize(pSize);
		}
	}

	/**
	 * Sets the horizontal gap between components.
	 * 
	 * @param hgap The new gap.
	 */
	public void setHGap(int hgap) {
		mHGap = hgap;
	}

	/**
	 * Sets the vertical gap between components.
	 * 
	 * @param vgap The new gap.
	 */
	public void setVGap(int vgap) {
		mVGap = vgap;
	}

	public void removeLayoutComponent(Component comp) {
		synchronized (comp.getTreeLock()) {
			if (comp == mCenter) {
				mCenter = null;
			} else if (comp == mNorth) {
				mNorth = null;
			} else if (comp == mSouth) {
				mSouth = null;
			} else if (comp == mEast) {
				mEast = null;
			} else if (comp == mWest) {
				mWest = null;
			}
		}
	}
}
