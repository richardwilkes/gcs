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

package com.trollworks.gcs.widgets.layout;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.Rectangle;

/** The basic unit within a {@link FlexLayout}. */
public abstract class FlexCell {
	private Alignment	mHorizontalAlignment	= Alignment.LEFT_TOP;
	private Alignment	mVerticalAlignment		= Alignment.CENTER;
	private Insets		mInsets					= new Insets(0, 0, 0, 0);
	private int			mX;
	private int			mY;
	private int			mWidth;
	private int			mHeight;

	/**
	 * Creates a new {@link FlexLayout} with this cell as its root cell and applies it to the
	 * specified component.
	 * 
	 * @param container The container to apply the {@link FlexLayout} to.
	 */
	public void apply(Container container) {
		container.setLayout(new FlexLayout(this));
	}

	/**
	 * Draws the borders of this cell. Useful for debugging.
	 * 
	 * @param gc The {@link Graphics} context to use.
	 * @param color The {@link Color} to use.
	 */
	public void draw(Graphics gc, Color color) {
		gc.setColor(color);
		gc.drawRect(mX, mY, mWidth, mHeight);
	}

	/**
	 * Layout the cell and its children.
	 * 
	 * @param bounds The bounds to use for the cell.
	 */
	public final void layout(Rectangle bounds) {
		bounds.x += mInsets.left;
		bounds.y += mInsets.top;
		bounds.width -= mInsets.left + mInsets.right;
		bounds.height -= mInsets.top + mInsets.bottom;
		mX = bounds.x;
		mY = bounds.y;
		mWidth = bounds.width;
		mHeight = bounds.height;
		layoutSelf(bounds);
	}

	/**
	 * Called to layout the cell and its children.
	 * 
	 * @param bounds The bounds to use for the cell. Insets have already been applied.
	 */
	protected abstract void layoutSelf(Rectangle bounds);

	/**
	 * @param type The type of size to determine.
	 * @return The size for this cell.
	 */
	public final Dimension getSize(LayoutSize type) {
		Dimension size = getSizeSelf(type);
		size.width += mInsets.left + mInsets.right;
		size.height += mInsets.top + mInsets.bottom;
		return LayoutSize.sanitizeSize(size);
	}

	/**
	 * @param type The type of size to determine.
	 * @return The size for this cell. Do not include the insets from the cell.
	 */
	protected abstract Dimension getSizeSelf(LayoutSize type);

	/** @return The horizontal alignment. */
	public Alignment getHorizontalAlignment() {
		return mHorizontalAlignment;
	}

	/** @param alignment The value to set for horizontal alignment. */
	public void setHorizontalAlignment(Alignment alignment) {
		mHorizontalAlignment = alignment;
	}

	/** @return The vertical alignment. */
	public Alignment getVerticalAlignment() {
		return mVerticalAlignment;
	}

	/** @param alignment The value to set for vertical alignment. */
	public void setVerticalAlignment(Alignment alignment) {
		mVerticalAlignment = alignment;
	}

	/** @return The insets. */
	public Insets getInsets() {
		return mInsets;
	}

	/** @param insets The value to set for insets. */
	public void setInsets(Insets insets) {
		if (insets != null) {
			mInsets.set(insets.top, insets.left, insets.bottom, insets.right);
		} else {
			mInsets.set(0, 0, 0, 0);
		}
	}

	@Override public String toString() {
		return getClass().getSimpleName();
	}
}
