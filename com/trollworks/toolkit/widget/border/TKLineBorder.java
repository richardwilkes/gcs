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

package com.trollworks.toolkit.widget.border;

import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Insets;

/** A border capable of displaying a line on one or more edges. */
public class TKLineBorder implements TKBorder {
	/** Constant specifying the left edge. */
	public static final int		LEFT_EDGE	= 1 << 0;
	/** Constant specifying the right edge. */
	public static final int		RIGHT_EDGE	= 1 << 1;
	/** Constant specifying the top edge. */
	public static final int		TOP_EDGE	= 1 << 2;
	/** Constant specifying the bottom edge. */
	public static final int		BOTTOM_EDGE	= 1 << 3;
	/** Constant specifying all edges. */
	public static final int		ALL_EDGES	= LEFT_EDGE | RIGHT_EDGE | TOP_EDGE | BOTTOM_EDGE;
	private static TKLineBorder	BLACK_LINE	= null;
	private static TKLineBorder	GRAY_LINE	= null;
	private Color				mLineColor;
	private int					mThickness;
	private int					mEdges;
	private boolean				mInsetBorder;

	/**
	 * Create a new line border.
	 * 
	 * @param edges The edges that require a border.
	 */
	public TKLineBorder(int edges) {
		this(Color.black, 1, edges, true);
	}

	/**
	 * Create a new line border.
	 * 
	 * @param color The line color to use.
	 */
	public TKLineBorder(Color color) {
		this(color, 1, ALL_EDGES, true);
	}

	/**
	 * Create a new line border.
	 * 
	 * @param color The line color to use.
	 * @param thickness The thickness, in pixels, of the line.
	 */
	public TKLineBorder(Color color, int thickness) {
		this(color, thickness, ALL_EDGES, true);
	}

	/**
	 * Create a new line border.
	 * 
	 * @param color The line color to use.
	 * @param thickness The thickness, in pixels, of the line.
	 * @param edges The edges that require a border.
	 */
	public TKLineBorder(Color color, int thickness, int edges) {
		this(color, thickness, edges, true);
	}

	/**
	 * Create a new line border.
	 * 
	 * @param color The line color to use.
	 * @param thickness The thickness, in pixels, of the line.
	 * @param edges The edges that require a border.
	 * @param insetBorder Pass in <code>true</code> if the border should report insets, otherwise
	 *            0-sized insets will be returned.
	 */
	public TKLineBorder(Color color, int thickness, int edges, boolean insetBorder) {
		mLineColor = color;
		mThickness = thickness;
		mEdges = edges;
		mInsetBorder = insetBorder;
	}

	/**
	 * @param isBlack Pass in <code>true</code> to get a black border, <code>false</code> to get
	 *            a gray border.
	 * @return A shared, 1-pixel thick border that surrounds all edges.
	 */
	public static TKLineBorder getSharedBorder(boolean isBlack) {
		if (isBlack) {
			if (BLACK_LINE == null) {
				BLACK_LINE = new TKLineBorder(Color.black);
			}
			return BLACK_LINE;
		}

		if (GRAY_LINE == null) {
			GRAY_LINE = new TKLineBorder(Color.gray);
		}
		return GRAY_LINE;
	}

	public void paintBorder(TKPanel panel, Graphics2D g2d, int x, int y, int width, int height) {
		Color oldColor = g2d.getColor();

		height--;
		width--;
		g2d.setColor(mLineColor);
		for (int i = 0; i < mThickness; i++) {
			if ((mEdges & LEFT_EDGE) != 0) {
				g2d.drawLine(x + i, y, x + i, y + height);
			}
			if ((mEdges & RIGHT_EDGE) != 0) {
				g2d.drawLine(x + width - i, y, x + width - i, y + height);
			}
			if ((mEdges & TOP_EDGE) != 0) {
				g2d.drawLine(x, y + i, x + width, y + i);
			}
			if ((mEdges & BOTTOM_EDGE) != 0) {
				g2d.drawLine(x, y + height - i, x + width, y + height - i);
			}
		}
		g2d.setColor(oldColor);
	}

	public Insets getBorderInsets(TKPanel panel) {
		return new Insets((mEdges & TOP_EDGE) != 0 && mInsetBorder ? mThickness : 0, (mEdges & LEFT_EDGE) != 0 && mInsetBorder ? mThickness : 0, (mEdges & BOTTOM_EDGE) != 0 && mInsetBorder ? mThickness : 0, (mEdges & RIGHT_EDGE) != 0 && mInsetBorder ? mThickness : 0);
	}
}
