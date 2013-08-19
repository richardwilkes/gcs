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

import java.awt.Dimension;
import java.awt.Rectangle;

/**
 * Components that want to have some control over scrolling when embedded in a {@link TKScrollPanel}
 * should implement this interface.
 */
public interface TKScrollable {
	/**
	 * Components that display logical rows or columns should compute the scroll increment that will
	 * completely expose one block of rows or columns, depending on the orientation.
	 * 
	 * @param visibleBounds The visible bounds of the scrolling area.
	 * @param vertical Whether the scroll is vertical or horizontal.
	 * @param upLeftDirection Whether the scroll is up/left or down/right.
	 * @return The amount to scroll.
	 */
	public int getBlockScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection);

	/** @return The preferred size of the viewport for a component. */
	public Dimension getPreferredViewportSize();

	/**
	 * Components that display logical rows or columns should compute the scroll increment that will
	 * completely expose one new row or column, depending on the orientation.
	 * 
	 * @param visibleBounds The visible bounds of the scrolling area.
	 * @param vertical Whether the scroll is vertical or horizontal.
	 * @param upLeftDirection Whether the scroll is up/left or down/right.
	 * @return The amount to scroll.
	 */
	public int getUnitScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection);

	/**
	 * @return <code>true</code> if a viewport should always force the height of this component to
	 *         match the height of the viewport.
	 */
	public boolean shouldTrackViewportHeight();

	/**
	 * @return <code>true</code> if a viewport should always force the width of this component to
	 *         match the width of the viewport.
	 */
	public boolean shouldTrackViewportWidth();
}
