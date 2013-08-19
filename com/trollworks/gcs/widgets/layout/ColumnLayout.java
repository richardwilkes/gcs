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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets.layout;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.LayoutManager2;

/**
 * Provides a standardized column layout. Columns are sized according to the preferred size of its
 * contents. Width is altered to accommodate available space while respecting minimum/maximum sizes,
 * while height is altered in in one three ways:
 * <ul>
 * <li>{@link RowDistribution#USE_PREFERRED_HEIGHT} - use the preferred height of each component.
 * <li>{@link RowDistribution#DISTRIBUTE_HEIGHT} - Give each component its preferred height,
 * distributing any excess height evenly among all components.
 * <li>{@link RowDistribution#GIVE_EXCESS_TO_LAST} - Give each component its preferred height, but
 * give any excess height to the last component in a column.
 * </ul>
 * This layout manager does not track any state specific to components, so it may be re-used in
 * multiple containers without creating new copies.
 */
public class ColumnLayout implements LayoutManager2 {
	/** The default horizontal gap size. */
	public static final int	DEFAULT_H_GAP_SIZE	= 5;
	/** The default vertical gap size. */
	public static final int	DEFAULT_V_GAP_SIZE	= 2;
	private int				mColumns;
	private int				mHGap;
	private int				mVGap;
	private RowDistribution	mDistribution;

	/** Creates a new layout with a single column. */
	public ColumnLayout() {
		this(1, DEFAULT_H_GAP_SIZE, DEFAULT_V_GAP_SIZE, RowDistribution.USE_PREFERRED_HEIGHT);
	}

	/**
	 * Creates a new layout with the specified number of columns.
	 * 
	 * @param columns The number of columns.
	 */
	public ColumnLayout(int columns) {
		this(columns, DEFAULT_H_GAP_SIZE, DEFAULT_V_GAP_SIZE, RowDistribution.USE_PREFERRED_HEIGHT);
	}

	/**
	 * Creates a new layout with the specified number of columns and height distribution.
	 * 
	 * @param columns The number of columns.
	 * @param distribution The height distribution style.
	 */
	public ColumnLayout(int columns, RowDistribution distribution) {
		this(columns, DEFAULT_H_GAP_SIZE, DEFAULT_V_GAP_SIZE, distribution);
	}

	/**
	 * Creates a new layout.
	 * 
	 * @param columns The number of columns.
	 * @param hgap The horizontal gap value.
	 * @param vgap The vertical gap value.
	 */
	public ColumnLayout(int columns, int hgap, int vgap) {
		this(columns, hgap, vgap, RowDistribution.USE_PREFERRED_HEIGHT);
	}

	/**
	 * Creates a new layout.
	 * 
	 * @param columns The number of columns.
	 * @param hgap The horizontal gap value.
	 * @param vgap The vertical gap value.
	 * @param distribution The height distribution style.
	 */
	public ColumnLayout(int columns, int hgap, int vgap, RowDistribution distribution) {
		if (columns < 1) {
			throw new IllegalArgumentException("columns must be greater than zero"); //$NON-NLS-1$
		}

		mColumns = columns;
		mHGap = hgap;
		mVGap = vgap;
		mDistribution = distribution;
	}

	public void addLayoutComponent(String name, Component component) {
		// Nothing to do...
	}

	public void addLayoutComponent(Component component, Object constraints) {
		// Nothing to do...
	}

	/** @return The way vertical space will be distributed to components. */
	public RowDistribution getHeightDistribution() {
		return mDistribution;
	}

	/** @return The horizontal gap used by this layout. */
	public int getHorizontalGap() {
		return mHGap;
	}

	/** @return The vertical gap used by this layout. */
	public int getVerticalGap() {
		return mVGap;
	}

	public float getLayoutAlignmentX(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return Component.CENTER_ALIGNMENT;
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public void layoutContainer(Container parent) {
		synchronized (parent.getTreeLock()) {
			Dimension pSize = parent.getSize();
			Insets insets = parent.getInsets();
			int compCount = parent.getComponentCount();
			int y = insets.top;
			int rows = 1 + (compCount - 1) / mColumns;
			int[] widths = new int[mColumns];
			int[] minWidths = new int[mColumns];
			int[] maxWidths = new int[mColumns];
			int[] heights = new int[rows];
			int width = pSize.width - (insets.left + insets.right + (mColumns - 1) * mHGap);
			int height = pSize.height - (insets.top + insets.bottom + (rows - 1) * mVGap);
			Dimension[] prefSizes = new Dimension[compCount];
			Dimension[] maxSizes = new Dimension[compCount];
			Dimension[] minSizes = new Dimension[compCount];
			Component comp;
			int i;
			int j;
			int k;
			int which;
			int portion;
			int participants;

			for (i = 0; i < compCount; i++) {
				comp = parent.getComponent(i);
				prefSizes[i] = comp.getPreferredSize();
				maxSizes[i] = comp.getMaximumSize();
				minSizes[i] = comp.getMinimumSize();
			}

			for (i = 0; i < compCount; i += mColumns) {
				for (j = i; j < i + mColumns && j < compCount; j++) {
					int compWidth = prefSizes[j].width;

					which = j - i;
					if (compWidth > widths[which]) {
						widths[which] = compWidth;
					}
					compWidth = minSizes[j].width;
					if (compWidth > minWidths[which]) {
						minWidths[which] = compWidth;
					}
					compWidth = maxSizes[j].width;
					if (compWidth > maxWidths[which]) {
						maxWidths[which] = compWidth;
					}
				}
			}

			for (i = 0; i < mColumns; i++) {
				width -= widths[i];
			}

			j = mColumns;
			if (width > 0) {
				while (width > 0 && j > 0) {
					portion = width / j;
					if (portion == 0) {
						portion = 1;
					}
					for (i = j = 0; i < mColumns && width != 0; i++) {
						if (widths[i] < maxWidths[i]) {
							if (widths[i] + portion <= maxWidths[i]) {
								widths[i] += portion;
								width -= portion;
							} else {
								width -= maxWidths[i] - widths[i];
								widths[i] = maxWidths[i];
							}
						}
						if (widths[i] < maxWidths[i]) {
							j++;
						}
						if (portion > width) {
							portion = width;
						}
					}
				}
			} else if (width < 0) {
				width = -width;
				while (width > 0 && j > 0) {
					portion = width / j;
					if (portion == 0) {
						portion = 1;
					}
					for (i = j = 0; i < mColumns && width != 0; i++) {
						if (widths[i] > minWidths[i]) {
							if (widths[i] - portion >= minWidths[i]) {
								widths[i] -= portion;
								width -= portion;
							} else {
								width -= widths[i] - minWidths[i];
								widths[i] = minWidths[i];
							}
						}
						if (widths[i] > minWidths[i]) {
							j++;
						}
						if (portion > width) {
							portion = width;
						}
					}
				}
			}

			portion = height;
			for (i = 0; i < rows; i++) {
				heights[i] = 0;
				for (j = 0; j < mColumns; j++) {
					k = i * mColumns + j;
					if (k < compCount) {
						k = prefSizes[k].height;
						if (k > heights[i]) {
							heights[i] = k;
						}
					}
				}
				portion -= heights[i];
			}

			participants = rows;
			while (portion < 0 && participants > 0) {
				int newPortion = participants > 1 ? 1 + -portion / participants : -portion;

				participants = 0;
				for (i = 0; i < rows; i++) {
					int newHeight = heights[i] - newPortion;
					boolean doit = true;

					for (j = 0; j < mColumns; j++) {
						k = i * mColumns + j;
						if (k < compCount) {
							int minHeight = minSizes[k].height;

							if (minHeight > newHeight) {
								if (minHeight < heights[i]) {
									newHeight = minHeight;
								} else {
									doit = false;
									break;
								}
							}
						}
					}

					if (doit) {
						portion += heights[i] - newHeight;
						heights[i] = newHeight;
						participants++;
					}
				}
			}

			if (portion > 0 && rows > 0) {
				if (mDistribution == RowDistribution.DISTRIBUTE_HEIGHT) {
					j = portion / rows;
					if (j > 0) {
						for (i = 0; i < rows; i++) {
							heights[i] += j;
						}
					}
					for (i = 0; i < portion - j * rows; i++) {
						heights[i]++;
					}
				} else if (mDistribution == RowDistribution.GIVE_EXCESS_TO_LAST) {
					heights[rows - 1] += portion;
				}
			}

			for (i = 0; i < rows; i++) {
				int x = insets.left;

				for (j = 0; j < mColumns; j++) {
					k = i * mColumns + j;
					if (k < compCount) {
						int compY = y;
						int cheight = heights[i];

						if (cheight < minSizes[k].height) {
							cheight = minSizes[k].height;
						}

						comp = parent.getComponent(k);

						if (cheight > maxSizes[k].height) {
							float align = comp.getAlignmentY();

							cheight = maxSizes[k].height;

							if (align >= 0.75) {
								compY += heights[i] - cheight;
							} else if (align >= 0.25) {
								compY += (heights[i] - cheight) / 2;
							}
						}
						comp.setBounds(x, compY, widths[j], cheight);
						x += widths[j] + mHGap;
					}
				}
				y += heights[i] + mVGap;
			}
		}
	}

	public Dimension maximumLayoutSize(Container parent) {
		long height = 0;
		long width = (mColumns - 1) * mHGap;

		synchronized (parent.getTreeLock()) {
			Insets insets = parent.getInsets();
			int compCount = parent.getComponentCount();
			int[] widths = new int[mColumns];
			int rows = 1 + (compCount - 1) / mColumns;
			int i;

			width += insets.left + insets.right;

			for (int y = 0; y < rows; y++) {
				int rowHeight = 0;

				for (int x = 0; x < mColumns; x++) {
					i = y * mColumns + x;
					if (i < compCount) {
						Dimension size = parent.getComponent(i).getMaximumSize();

						if (size.height > rowHeight) {
							rowHeight = size.height;
						}
						if (size.width > widths[x]) {
							widths[x] = size.width;
						}
					}
				}
				height += rowHeight;
			}
			height += insets.top + insets.bottom;
			if (rows > 0) {
				height += (rows - 1) * mVGap;
			}

			for (i = 0; i < mColumns; i++) {
				width += widths[i];
			}
		}

		if (width > LayoutSize.MAXIMUM_SIZE) {
			width = LayoutSize.MAXIMUM_SIZE;
		}
		if (height > LayoutSize.MAXIMUM_SIZE) {
			height = LayoutSize.MAXIMUM_SIZE;
		}

		return new Dimension((int) width, (int) height);
	}

	public Dimension minimumLayoutSize(Container parent) {
		int height = 0;
		int width = (mColumns - 1) * mHGap;

		synchronized (parent.getTreeLock()) {
			Insets insets = parent.getInsets();
			int compCount = parent.getComponentCount();
			int[] widths = new int[mColumns];
			int rows = 1 + (compCount - 1) / mColumns;
			int i;

			width += insets.left + insets.right;

			for (int y = 0; y < rows; y++) {
				int rowHeight = 0;

				for (int x = 0; x < mColumns; x++) {
					i = y * mColumns + x;
					if (i < compCount) {
						Dimension size = parent.getComponent(i).getMinimumSize();

						if (size.width > widths[x]) {
							widths[x] = size.width;
						}
						if (size.height > rowHeight) {
							rowHeight = size.height;
						}
					}
				}
				height += rowHeight;
			}
			height += insets.top + insets.bottom;
			if (rows > 0) {
				height += (rows - 1) * mVGap;
			}

			for (i = 0; i < mColumns; i++) {
				width += widths[i];
			}
		}
		return new Dimension(width, height);
	}

	public Dimension preferredLayoutSize(Container parent) {
		int height = 0;
		int width = (mColumns - 1) * mHGap;

		synchronized (parent.getTreeLock()) {
			Insets insets = parent.getInsets();
			int compCount = parent.getComponentCount();
			int[] widths = new int[mColumns];
			int rows = 1 + (compCount - 1) / mColumns;
			int i;

			width += insets.left + insets.right;

			for (int y = 0; y < rows; y++) {
				int rowHeight = 0;

				for (int x = 0; x < mColumns; x++) {
					i = y * mColumns + x;
					if (i < compCount) {
						Dimension size = parent.getComponent(i).getPreferredSize();

						if (size.height > rowHeight) {
							rowHeight = size.height;
						}
						if (size.width > widths[x]) {
							widths[x] = size.width;
						}
					}
				}
				height += rowHeight;
			}
			height += insets.top + insets.bottom;
			if (rows > 0) {
				height += (rows - 1) * mVGap;
			}

			for (i = 0; i < mColumns; i++) {
				width += widths[i];
			}
		}
		return new Dimension(width, height);
	}

	public void removeLayoutComponent(Component comp) {
		// Nothing to do...
	}
}
