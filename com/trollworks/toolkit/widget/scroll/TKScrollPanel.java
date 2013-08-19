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

import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKReshapeListener;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;

import java.awt.AWTEvent;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseWheelEvent;
import java.awt.geom.Rectangle2D;

/** A panel that can scroll its contents. */
public class TKScrollPanel extends TKPanel implements ActionListener, TKScrollBarOwner, TKReshapeListener, LayoutManager2, Runnable {
	/** Constant representing horizontal scroll bar only. */
	public static final int		HORIZONTAL			= -1;
	/** Constant representing both horizontal and vertical scroll bars. */
	public static final int		BOTH				= 0;
	/** Constant representing vertical scroll bar only. */
	public static final int		VERTICAL			= 1;
	/** The auto-scrolling delay. */
	public static final int		AUTOSCROLL_DELAY	= 75;
	private TKPanel				mBottomLeftCorner;
	private TKPanel				mBottomRightCorner;
	private TKPanel				mTopLeftCorner;
	private TKPanel				mTopRightCorner;
	private TKPanel				mVerticalHeader;
	private TKPanel				mHorizontalHeader;
	private TKScrollContentView	mContentView;
	private TKPanel				mContentBorderView;
	private TKScrollContentView	mVerticalHeaderView;
	private TKScrollContentView	mHorizontalHeaderView;
	private TKScrollBar			mHorizontalScrollBar;
	private TKScrollBar			mVerticalScrollBar;
	private boolean				mBottomLeftIsDynamic;
	private boolean				mIgnoreScroll;
	private boolean				mAutoScrolls;
	private TKPanel				mAutoScrollComponent;
	private int					mAutoScrollX;
	private int					mAutoScrollY;
	private int					mAutoScrollModifiers;

	/** Create a new scroll panel with both scroll bars visible and no content. */
	public TKScrollPanel() {
		this(BOTH, null);
	}

	/**
	 * Create a new scroll panel.
	 * 
	 * @param type One of {@link #HORIZONTAL}, {@link #VERTICAL}, or {@link #BOTH}.
	 */
	public TKScrollPanel(int type) {
		this(type, null);
	}

	/**
	 * Create a new scroll panel with both scroll bars visible.
	 * 
	 * @param panel The panel to embed as the content.
	 */
	public TKScrollPanel(TKPanel panel) {
		this(BOTH, panel);
	}

	/**
	 * Create a new scroll panel.
	 * 
	 * @param type One of {@link #HORIZONTAL}, {@link #VERTICAL}, or {@link #BOTH}.
	 * @param panel The panel to embed as the content.
	 */
	public TKScrollPanel(int type, TKPanel panel) {
		super();

		mAutoScrolls = true;

		setTopLeftCorner(createDefaultCorner());
		setTopRightCorner(createDefaultCorner());
		setBottomLeftCorner(createDefaultCorner());
		setBottomRightCorner(createDefaultCorner());

		mContentBorderView = new TKPanel();
		mContentBorderView.setLayout(new TKCompassLayout());
		add(mContentBorderView);

		mContentView = new TKScrollContentView();
		mContentView.addReshapeListener(this);
		mContentBorderView.add(mContentView, TKCompassPosition.CENTER);

		if (type <= BOTH) {
			mHorizontalScrollBar = new TKScrollBar(this, false, 0, 100, 0);
			mHorizontalScrollBar.addActionListener(this);
			add(mHorizontalScrollBar);
		}
		if (type >= BOTH) {
			mVerticalScrollBar = new TKScrollBar(this, true, 0, 100, 0);
			mVerticalScrollBar.addActionListener(this);
			add(mVerticalScrollBar);
		}

		setLayout(this);
		setContent(panel);
		enableAWTEvents(AWTEvent.MOUSE_WHEEL_EVENT_MASK);
	}

	@Override public boolean isValidateRoot() {
		return true;
	}

	public void actionPerformed(ActionEvent event) {
		TKPanel content = mContentView.getContent();

		if (content != null) {
			int ox = content.getX();
			int oy = content.getY();
			int nx = mHorizontalScrollBar != null ? -mHorizontalScrollBar.getCurrentValue() : ox;
			int ny = mVerticalScrollBar != null ? -mVerticalScrollBar.getCurrentValue() : oy;

			if (ox != nx || oy != ny) {
				scrollDelta(nx - ox, ny - oy);
			}
		}
	}

	public void addLayoutComponent(String name, Component component) {
		// Nothing to do...
	}

	public void addLayoutComponent(Component component, Object constraints) {
		// Nothing to do...
	}

	public void panelReshaped(TKPanel component, boolean moved, boolean resized) {
		if (component == mContentView || component == mContentView.getContent()) {
			syncScrollBars();
		} else if (isBottomLeftIsDynamic() && component == mBottomLeftCorner) {
			revalidate();
		}
	}

	/** @return A new default corner object. */
	protected TKPanel createDefaultCorner() {
		TKPanel corner = new TKPanel(new FlowLayout(FlowLayout.LEFT, 0, 0));

		corner.setOpaque(true);
		return corner;
	}

	/**
	 * @return <code>true</code> if this scroll pane auto-scrolls when
	 *         <code>scrollPointIntoView</code> is called.
	 */
	public boolean getAutoScrolls() {
		return mAutoScrolls;
	}

	public int getBlockScrollIncrement(boolean vertical, boolean upLeftDirection) {
		TKPanel content = mContentView.getContent();
		Rectangle bounds = mContentView.getScrollingBounds(true);

		if (content instanceof TKScrollable) {
			Point pt = new Point();

			convertPoint(pt, mContentView, content);
			bounds.x = pt.x;
			bounds.y = pt.y;
			Rectangle2D.intersect(content.getLocalBounds(), bounds, bounds);
			return ((TKScrollable) content).getBlockScrollIncrement(bounds, vertical, upLeftDirection);
		}

		return (upLeftDirection ? -1 : 1) * (vertical ? bounds.height : bounds.width);
	}

	/** @return The panel to be displayed in the bottom, left corner. */
	public TKPanel getBottomLeftCorner() {
		return mBottomLeftCorner;
	}

	/** @return The panel to be displayed in the bottom, right corner. */
	public TKPanel getBottomRightCorner() {
		return mBottomRightCorner;
	}

	public TKPanel getContentBorderView() {
		return mContentBorderView;
	}

	public int getContentSize(boolean vertical) {
		TKPanel content = mContentView.getContent();

		if (content != null) {
			return vertical ? content.getHeight() : content.getWidth();
		}
		return 0;
	}

	public TKScrollContentView getContentView() {
		return mContentView;
	}

	public int getContentViewSize(boolean vertical) {
		return vertical ? mContentView.getHeight() : mContentView.getWidth();
	}

	public float getLayoutAlignmentX(Container target) {
		return CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return CENTER_ALIGNMENT;
	}

	/** @return The panel to be displayed in the top, left corner. */
	public TKPanel getTopLeftCorner() {
		return mTopLeftCorner;
	}

	/** @return The panel to be displayed in the top, right corner. */
	public TKPanel getTopRightCorner() {
		return mTopRightCorner;
	}

	public int getUnitScrollIncrement(boolean vertical, boolean upLeftDirection) {
		TKPanel content = mContentView.getContent();

		if (content instanceof TKScrollable) {
			Point pt = new Point();
			Rectangle bounds = mContentView.getScrollingBounds(true);

			convertPoint(pt, mContentView, content);
			bounds.x = pt.x;
			bounds.y = pt.y;
			Rectangle2D.intersect(content.getLocalBounds(), bounds, bounds);
			return ((TKScrollable) content).getUnitScrollIncrement(bounds, vertical, upLeftDirection);
		}

		return upLeftDirection ? -1 : 1;
	}

	/**
	 * @param vertical Whether the scroll bar is vertical or horizontal.
	 * @return The specified scroll bar, or <code>null</code> if there isn't one.
	 */
	public TKScrollBar getScrollBar(boolean vertical) {
		return vertical ? mVerticalScrollBar : mHorizontalScrollBar;
	}

	/** @return The current position of the view within the content. */
	public Point getViewPosition() {
		TKPanel content = mContentView.getContent();

		if (content != null) {
			return new Point(-content.getX(), -content.getY());
		}
		return new Point();
	}

	/**
	 * @return <code>true</code> if the bottom, left corner will expand horizontally to
	 *         accommodate its contents.
	 */
	public boolean isBottomLeftIsDynamic() {
		return mBottomLeftIsDynamic;
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public void layoutContainer(Container target) {
		Insets insets = getInsets();
		int x = insets.left;
		int y = insets.top;
		int width = getWidth() - (x + insets.right);
		int height = getHeight() - (y + insets.bottom);
		int hHeight = 0;
		int hWidth = 0;
		int sHeight = 0;
		int sWidth = 0;

		if (mHorizontalHeaderView != null) {
			hHeight = mHorizontalHeader.getPreferredSize().height;
		}

		if (mVerticalHeaderView != null) {
			hWidth = mVerticalHeader.getPreferredSize().width;
		}

		if (mHorizontalScrollBar != null) {
			sHeight = mHorizontalScrollBar.getPreferredSize().height;
		}

		if (mVerticalScrollBar != null) {
			sWidth = mVerticalScrollBar.getPreferredSize().width;
		}

		if (mTopLeftCorner != null) {
			mTopLeftCorner.setBounds(x, y, hWidth, hHeight);
		}

		if (mTopRightCorner != null) {
			mTopRightCorner.setBounds(x + width - sWidth, y, sWidth, hHeight);
		}

		if (mBottomRightCorner != null) {
			mBottomRightCorner.setBounds(x + width - sWidth, y + height - sHeight, sWidth, sHeight);
		}

		if (mHorizontalHeaderView != null) {
			mHorizontalHeaderView.setBounds(x + hWidth, y, width - (hWidth + sWidth), hHeight);
		}

		if (mVerticalHeaderView != null) {
			mVerticalHeaderView.setBounds(x, y + hHeight, hWidth, height - (hHeight + sHeight));
		}

		if (mVerticalScrollBar != null) {
			if (mHorizontalScrollBar == null) {
				mVerticalScrollBar.setBounds(x + width - sWidth, y + hHeight, sWidth, height - (hHeight + (mBottomRightCorner != null ? sHeight : 0)));
			} else {
				mVerticalScrollBar.setBounds(x + width - sWidth, y + hHeight, sWidth, height - (hHeight + sHeight));
			}
		}

		mContentBorderView.setBounds(x + hWidth, y + hHeight, width - (hWidth + sWidth), height - (hHeight + sHeight));

		if (mBottomLeftCorner != null) {
			if (isBottomLeftIsDynamic()) {
				int tmp = mBottomLeftCorner.getPreferredSize().width;

				if (tmp > hWidth) {
					hWidth = tmp;
				}
				tmp = width - (hWidth + sWidth);
				if (tmp < 32) {
					hWidth -= 32 - tmp;
				}
			}
			mBottomLeftCorner.setBounds(x, y + height - sHeight, hWidth, sHeight);
		}

		if (mHorizontalScrollBar != null) {
			mHorizontalScrollBar.setBounds(x + hWidth, y + height - sHeight, width - (hWidth + sWidth), sHeight);
		}
	}

	public Dimension maximumLayoutSize(Container target) {
		return new Dimension(MAX_SIZE, MAX_SIZE);
	}

	public Dimension minimumLayoutSize(Container target) {
		Insets insets = getInsets();
		Dimension size = mContentBorderView.getMinimumSize();

		if (mHorizontalHeaderView != null) {
			size.height += mHorizontalHeader.getPreferredSize().height;
		}

		if (mVerticalHeaderView != null) {
			size.width += mVerticalHeader.getPreferredSize().width;
		}

		if (mHorizontalScrollBar != null) {
			size.height += mHorizontalScrollBar.getPreferredSize().height;
		}

		if (mVerticalScrollBar != null) {
			size.width += mVerticalScrollBar.getPreferredSize().width;
		}

		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom;

		return size;
	}

	public Dimension preferredLayoutSize(Container target) {
		Insets insets = mContentBorderView.getInsets();
		TKPanel content = mContentView.getContent();
		Dimension size;

		if (content != null) {
			if (content instanceof TKScrollable) {
				size = ((TKScrollable) content).getPreferredViewportSize();
			} else {
				size = content.getPreferredSize();
			}
		} else {
			size = new Dimension();
		}

		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom;

		if (mHorizontalHeaderView != null) {
			size.height += mHorizontalHeader.getPreferredSize().height;
		}

		if (mVerticalHeaderView != null) {
			size.width += mVerticalHeader.getPreferredSize().width;
		}

		if (mHorizontalScrollBar != null) {
			size.height += mHorizontalScrollBar.getPreferredSize().height;
		}

		if (mVerticalScrollBar != null) {
			size.width += mVerticalScrollBar.getPreferredSize().width;
		}

		insets = getInsets();
		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom;

		return size;
	}

	public void removeLayoutComponent(Component target) {
		// Nothing to do...
	}

	public void run() {
		TKPanel component = mAutoScrollComponent;

		mAutoScrollComponent = null;
		if (component != null) {
			component.processMouseMotionEvent(new TKAutoScrollEvent(component, mAutoScrollModifiers, mAutoScrollX, mAutoScrollY));
		}
	}

	/**
	 * Scroll.
	 * 
	 * @param upperLeft Pass in <code>true</code> to scroll to the upper-left corner or
	 *            <code>false</code> to scroll to the bottom-right corner.
	 */
	public void scroll(boolean upperLeft) {
		if (mVerticalScrollBar != null) {
			mVerticalScrollBar.setCurrentValue(upperLeft ? 0 : mVerticalScrollBar.getMaximumValue());
		}
		if (mHorizontalScrollBar != null) {
			mHorizontalScrollBar.setCurrentValue(upperLeft ? 0 : mHorizontalScrollBar.getMaximumValue());
		}
	}

	public void scroll(boolean vertical, boolean upLeftDirection, boolean page) {
		if (vertical) {
			if (mVerticalScrollBar != null) {
				mVerticalScrollBar.scroll(upLeftDirection, page);
			}
		} else {
			if (mHorizontalScrollBar != null) {
				mHorizontalScrollBar.scroll(upLeftDirection, page);
			}
		}
	}

	/**
	 * Scroll.
	 * 
	 * @param x The x-coordinate to scroll to.
	 * @param y The y-coordinate to scroll to.
	 */
	public void scroll(int x, int y) {
		x = mHorizontalScrollBar != null ? -(x - mHorizontalScrollBar.getCurrentValue()) : 0;
		y = mVerticalScrollBar != null ? -(y - mVerticalScrollBar.getCurrentValue()) : 0;
		scrollDelta(x, y);
	}

	private void scrollDelta(int dx, int dy) {
		if (!mIgnoreScroll && (dx != 0 || dy != 0)) {
			TKPanel content = mContentView.getContent();

			if (content != null) {
				int ox = content.getX();
				int oy = content.getY();
				int nx = ox + dx;
				int ny = oy + dy;

				mIgnoreScroll = true;
				if (dx != 0 && mHorizontalScrollBar != null) {
					mHorizontalScrollBar.setCurrentValue(-nx);
					nx = -mHorizontalScrollBar.getCurrentValue();
				} else {
					dx = 0;
					nx = ox;
				}

				if (dy != 0 && mVerticalScrollBar != null) {
					mVerticalScrollBar.setCurrentValue(-ny);
					ny = -mVerticalScrollBar.getCurrentValue();
				} else {
					dy = 0;
					ny = oy;
				}

				mIgnoreScroll = false;

				if (ox != nx || oy != ny) {
					content.setLocation(nx, ny);

					if (ox != nx && mHorizontalHeader != null) {
						mHorizontalHeader.setLocation(nx, 0);
					}

					if (oy != ny && mVerticalHeader != null) {
						mVerticalHeader.setLocation(0, ny);
					}
				}
			}
		}
	}

	public Point scrollPointIntoView(MouseEvent event, TKPanel panel, Point viewPoint) {
		TKPanel content = mContentView.getContent();
		Point deltaPt = new Point();
		Point pt = new Point(viewPoint);

		if (pt.x < 0) {
			pt.x = 0;
		} else if (pt.x >= panel.getWidth()) {
			pt.x = panel.getWidth() - 1;
		}

		if (pt.y < 0) {
			pt.y = 0;
		} else if (pt.y >= panel.getHeight()) {
			pt.y = panel.getHeight() - 1;
		}

		if (panel != content) {
			Container parent = panel.getParent();

			while (parent != null && parent != content) {
				parent = parent.getParent();
			}

			if (parent != content) {
				return deltaPt;
			}

			convertPoint(pt, panel, content);
		}

		if (pt.x >= 0 && pt.x < content.getWidth() && pt.y >= 0 && pt.y < content.getHeight()) {
			Point viewPosition = getViewPosition();
			Point newPosition;

			content.scrollRectIntoView(new Rectangle(pt.x, pt.y, 1, 1));
			newPosition = getViewPosition();

			deltaPt.x = newPosition.x - viewPosition.x;
			deltaPt.y = newPosition.y - viewPosition.y;
			if (deltaPt.x != 0) {
				deltaPt.x += deltaPt.x < 0 ? -6 : 6;
			}
			if (deltaPt.y != 0) {
				deltaPt.y += deltaPt.y < 0 ? -6 : 6;
			}
			if (getAutoScrolls() && (deltaPt.x != 0 || deltaPt.y != 0)) {
				mAutoScrollX = event.getX() + deltaPt.x;
				mAutoScrollY = event.getY() + deltaPt.y;
				mAutoScrollModifiers = event.getModifiers();
				if (mAutoScrollComponent == null) {
					mAutoScrollComponent = panel;
					TKTimerTask.schedule(this, AUTOSCROLL_DELAY);
				}
			} else {
				mAutoScrollComponent = null;
			}
		}

		return deltaPt;
	}

	/**
	 * Scrolls the specified bounds into view.
	 * 
	 * @param bounds The bounds to scroll into view.
	 */
	@Override public void scrollRectIntoView(Rectangle bounds) {
		TKPanel content = mContentView.getContent();

		if (content != null) {
			Rectangle vBounds = mContentView.getScrollingBounds(false);

			if (!vBounds.isEmpty() && !vBounds.contains(bounds)) {
				int savedX = bounds.x;
				int savedY = bounds.y;
				int savedWidth = bounds.width;
				int savedHeight = bounds.height;
				int dx = 0;
				int dy = 0;
				int tmp;

				if (bounds.width >= vBounds.width && !bounds.intersects(vBounds)) {
					bounds.width = vBounds.width - 1;
				}

				if (bounds.width < vBounds.width) {
					tmp = vBounds.x + vBounds.width;
					if (bounds.x >= tmp) {
						dx = tmp - bounds.width;
					}
					if (bounds.x + dx + bounds.width > tmp) {
						dx -= bounds.x + dx + bounds.width - tmp;
					}
					if (bounds.x + dx < vBounds.x) {
						dx += vBounds.x - (bounds.x + dx);
					}
				}

				if (bounds.height >= vBounds.height && !bounds.intersects(vBounds)) {
					bounds.height = vBounds.height - 1;
				}

				if (bounds.height < vBounds.height) {
					tmp = vBounds.y + vBounds.height;
					if (bounds.y >= tmp) {
						dy = tmp - bounds.height;
					}
					if (bounds.y + dy + bounds.height > tmp) {
						dy -= bounds.y + dy + bounds.height - tmp;
					}
					if (bounds.y + dy < vBounds.y) {
						dy += vBounds.y - (bounds.y + dy);
					}
				}

				convertRectangle(bounds, this, content);
				scrollDelta(dx, dy);
				bounds.x = savedX;
				bounds.y = savedY;
				bounds.width = savedWidth;
				bounds.height = savedHeight;
			}
		}

		super.scrollRectIntoView(bounds);
	}

	/** @param autoScrolls Whether to auto-scroll or not. */
	public void setAutoScrolls(boolean autoScrolls) {
		mAutoScrolls = autoScrolls;
	}

	/** @param corner The panel to be displayed in the bottom, left corner. */
	public void setBottomLeftCorner(TKPanel corner) {
		if (mBottomLeftCorner != null) {
			mBottomLeftCorner.removeReshapeListener(this);
			remove(mBottomLeftCorner);
		}
		mBottomLeftCorner = corner;
		if (mBottomLeftCorner != null) {
			add(mBottomLeftCorner);
			mBottomLeftCorner.addReshapeListener(this);
		}
		revalidate();
	}

	/**
	 * @param dynamic Whether the bottom, left corner should expand horizontally to accommodate its
	 *            contents.
	 */
	public void setBottomLeftIsDynamic(boolean dynamic) {
		mBottomLeftIsDynamic = dynamic;
		if (mBottomLeftCorner != null) {
			mBottomLeftCorner.addReshapeListener(this);
		}
		revalidate();
	}

	/** @param corner The panel to be displayed in the bottom, right corner. */
	public void setBottomRightCorner(TKPanel corner) {
		if (mBottomRightCorner != null) {
			remove(mBottomRightCorner);
		}
		mBottomRightCorner = corner;
		if (mBottomRightCorner != null) {
			add(mBottomRightCorner);
		}
		revalidate();
	}

	/** @return The content this scroll panel contains. */
	public TKPanel getContent() {
		return mContentView.getContent();
	}

	/** @param content The panel to be displayed as the content. */
	public void setContent(TKPanel content) {
		TKPanel oldContent = mContentView.getContent();

		if (oldContent != null) {
			oldContent.removeReshapeListener(this);
		}
		mContentView.setContent(content);
		if (content != null) {
			int x = mHorizontalScrollBar != null ? mHorizontalScrollBar.getCurrentValue() : 0;
			int y = mVerticalScrollBar != null ? mVerticalScrollBar.getCurrentValue() : 0;
			Dimension size;

			if (content instanceof TKScrollable) {
				TKScrollable scrollable = (TKScrollable) content;

				size = scrollable.getPreferredViewportSize();
				if (scrollable.shouldTrackViewportHeight()) {
					size.height = mContentView.getHeight();
				}
				if (scrollable.shouldTrackViewportWidth()) {
					size.width = mContentView.getWidth();
				}
			} else {
				size = content.getPreferredSize();
			}

			content.setBounds(-x, -y, size.width, size.height);
			content.addReshapeListener(this);
			syncScrollBars();
		}
		revalidate();
	}

	/** @param header The panel to be displayed as the horizontal header. */
	public void setHorizontalHeader(TKPanel header) {
		if (header != null && mHorizontalHeaderView == null) {
			mHorizontalHeaderView = new TKScrollContentView(mContentView);
			add(mHorizontalHeaderView);
		} else if (header == null && mHorizontalHeaderView != null) {
			remove(mHorizontalHeaderView);
		}

		mHorizontalHeader = header;
		if (mHorizontalHeader != null) {
			Dimension size = mHorizontalHeader.getPreferredSize();

			mHorizontalHeader.setBounds(0, 0, size.width, size.height);
			mHorizontalHeaderView.setContent(mHorizontalHeader);
		}
		revalidate();
	}

	/** @param corner The panel to be displayed in the top, left corner. */
	public void setTopLeftCorner(TKPanel corner) {
		if (mTopLeftCorner != null) {
			remove(mTopLeftCorner);
		}
		mTopLeftCorner = corner;
		if (mTopLeftCorner != null) {
			add(mTopLeftCorner);
		}
		revalidate();
	}

	/** @param corner The panel to be displayed in the top, right corner. */
	public void setTopRightCorner(TKPanel corner) {
		if (mTopRightCorner != null) {
			remove(mTopRightCorner);
		}
		mTopRightCorner = corner;
		if (mTopRightCorner != null) {
			add(mTopRightCorner);
		}
		revalidate();
	}

	/** @param header The panel to be displayed as the vertical header. */
	public void setVerticalHeader(TKPanel header) {
		if (header != null && mVerticalHeaderView == null) {
			mVerticalHeaderView = new TKScrollContentView(mContentView);
			add(mVerticalHeaderView);
		} else if (header == null && mVerticalHeaderView != null) {
			remove(mVerticalHeaderView);
		}

		mVerticalHeader = header;
		if (mVerticalHeader != null) {
			Dimension size = mVerticalHeader.getPreferredSize();

			mVerticalHeader.setBounds(0, 0, size.width, size.height);
			mVerticalHeaderView.setContent(mVerticalHeader);
		}
		revalidate();
	}

	public void stopAutoScroll() {
		mAutoScrollComponent = null;
	}

	/** Synchronizes the scroll bars with the current contents. */
	protected void syncScrollBars() {
		TKPanel content = mContentView.getContent();
		int max;

		if (content instanceof TKScrollable) {
			TKScrollable scrollable = (TKScrollable) content;
			Dimension size = content.getSize();
			Dimension vSize = mContentView.getSize();
			boolean needResize = false;

			if (scrollable.shouldTrackViewportHeight() && size.height != vSize.height) {
				size.height = vSize.height;
				needResize = true;
			}
			if (scrollable.shouldTrackViewportWidth() && size.width != vSize.width) {
				size.width = vSize.width;
				needResize = true;
			}
			if (needResize) {
				content.setSize(size);
				// We can return, since the resize will cause this routine to be called again.
				return;
			}
		}

		if (mHorizontalScrollBar != null) {
			max = getContentSize(false) - getContentViewSize(false);
			mHorizontalScrollBar.setMinimumValue(0);
			mHorizontalScrollBar.setMaximumValue(max < 0 ? 0 : max);
		}

		if (mVerticalScrollBar != null) {
			max = getContentSize(true) - getContentViewSize(true);
			mVerticalScrollBar.setMinimumValue(0);
			mVerticalScrollBar.setMaximumValue(max < 0 ? 0 : max);
		}
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		if (event.getID() == MouseEvent.MOUSE_WHEEL) {
			MouseWheelEvent mwe = (MouseWheelEvent) event;
			int type = mwe.getScrollType();
			boolean upLeft = mwe.getWheelRotation() > 0;
			boolean vertical = !mwe.isShiftDown();

			if (type == MouseWheelEvent.WHEEL_UNIT_SCROLL) {
				int amt = mwe.getUnitsToScroll();

				if (amt < 0) {
					amt = -amt;
				}
				if (vertical) {
					scrollDelta(0, getUnitScrollIncrement(true, upLeft) * amt);
				} else {
					scrollDelta(getUnitScrollIncrement(false, upLeft) * amt, 0);
				}
			} else if (type == MouseWheelEvent.WHEEL_BLOCK_SCROLL) {
				if (vertical) {
					scrollDelta(0, getBlockScrollIncrement(true, upLeft));
				} else {
					scrollDelta(getBlockScrollIncrement(false, upLeft), 0);
				}
			}
			mwe.consume();
		}
	}

	/**
	 * This is a static helper method for the common case of wanting a panel to be no smaller than
	 * the scroll view its contained in.
	 * 
	 * @param panel The panel to determine {@link TKScrollable#shouldTrackViewportHeight()} for.
	 * @return <code>true</code> if a viewport should always force the height of the specified
	 *         panel to match the height of the viewport.
	 */
	public static boolean shouldTrackViewportHeight(TKPanel panel) {
		TKScrollPanel scroller = (TKScrollPanel) panel.getAncestorOfType(TKScrollPanel.class);
		int height = panel.getPreferredSize().height;

		return (scroller != null ? scroller.getContentViewSize(true) : height) > height;
	}

	/**
	 * This is a static helper method for the common case of wanting a panel to be no smaller than
	 * the scroll view its contained in.
	 * 
	 * @param panel The panel to determine {@link TKScrollable#shouldTrackViewportWidth()} for.
	 * @return <code>true</code> if a viewport should always force the width of the specified
	 *         panel to match the width of the viewport.
	 */
	public static boolean shouldTrackViewportWidth(TKPanel panel) {
		TKScrollPanel scroller = (TKScrollPanel) panel.getAncestorOfType(TKScrollPanel.class);
		int width = panel.getPreferredSize().width;

		return (scroller != null ? scroller.getContentViewSize(false) : width) > width;
	}
}
