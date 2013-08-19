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

import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Point;
import java.awt.event.MouseEvent;

/**
 * Components that provide a scrolling view, such as {@link TKScrollPanel}, should implement this
 * interface to interact with their {@link TKScrollBar}s.
 */
public interface TKScrollBarOwner {
	/**
	 * @param vertical Whether the scroll is vertical or horizontal.
	 * @param upLeftDirection Whether the scroll is up/left or down/right.
	 * @return The unit increment.
	 */
	public int getUnitScrollIncrement(boolean vertical, boolean upLeftDirection);

	/**
	 * @param vertical Whether the scroll is vertical or horizontal.
	 * @param upLeftDirection Whether the scroll is up/left or down/right.
	 * @return The block increment.
	 */
	public int getBlockScrollIncrement(boolean vertical, boolean upLeftDirection);

	/**
	 * @param vertical Whether the scroll is vertical or horizontal.
	 * @return The size of the content in the specified orientation.
	 */
	public int getContentSize(boolean vertical);

	/**
	 * @param vertical Whether the scroll is vertical or horizontal.
	 * @return The size of the content view port in the specified orientation.
	 */
	public int getContentViewSize(boolean vertical);

	/** @return The content view port. */
	public TKScrollContentView getContentView();

	/** @return The content border view port. */
	public TKPanel getContentBorderView();

	/**
	 * Scroll.
	 * 
	 * @param vertical Pass in <code>true</code> to use the vertical scroll bar or
	 *            <code>false</code> to use the horizontal scroll bar.
	 * @param upLeftDirection Pass in <code>true</code> to scroll in the up or left direction,
	 *            <code>false</code> for the down or right direction.
	 * @param page Pass in <code>true</code> to scroll a full block at a time, <code>false</code>
	 *            to scroll one unit.
	 */
	public void scroll(boolean vertical, boolean upLeftDirection, boolean page);

	/**
	 * Scrolls the specified point into view and initiates auto-scrolling if that has been enabled.
	 * This method should be called during mouse drag operations.
	 * 
	 * @param event The mouse event to react to.
	 * @param panel The panel.
	 * @param viewPoint The point within the panel that should be kept in view.
	 * @return The amount scrolled.
	 */
	public Point scrollPointIntoView(MouseEvent event, TKPanel panel, Point viewPoint);

	/** Stops any pending auto-scroll. */
	public void stopAutoScroll();
}
