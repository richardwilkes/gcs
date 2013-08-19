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

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.utility.Colors;
import com.trollworks.gcs.utility.Debug;
import com.trollworks.gcs.utility.Geometry;
import com.trollworks.gcs.utility.io.DragUtil;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.widgets.ActionPanel;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.GraphicsUtilities;
import com.trollworks.gcs.widgets.UIUtilities;

import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Composite;
import java.awt.Container;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Shape;
import java.awt.Transparency;
import java.awt.dnd.Autoscroll;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DragGestureEvent;
import java.awt.dnd.DragGestureListener;
import java.awt.dnd.DragSource;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.ActionEvent;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.StringTokenizer;

import javax.swing.JScrollPane;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;
import javax.swing.ToolTipManager;
import javax.swing.UIManager;
import javax.swing.undo.StateEdit;
import javax.swing.undo.UndoManager;
import javax.swing.undo.UndoableEdit;

/** A panel that can show both hierarchical and tabular data. */
public class Outline extends ActionPanel implements OutlineModelListener, ComponentListener, FocusListener, Autoscroll, Scrollable, DragGestureListener, DropTargetListener, MouseListener, MouseMotionListener, KeyListener {
	private static String			MSG_SORT_UNDO_TITLE;
	private static String			MSG_ROW_DROP_UNDO_TITLE;
	/** The default double-click action command. */
	public static final String		CMD_OPEN_SELECTION					= "Outline.OpenSelection";				//$NON-NLS-1$
	/** The default selection changed action command. */
	public static final String		CMD_SELECTION_CHANGED				= "Outline.SelectionChanged";			//$NON-NLS-1$
	/** The default potential content size change action command. */
	public static final String		CMD_POTENTIAL_CONTENT_SIZE_CHANGE	= "Outline.ContentSizeMayHaveChanged";	//$NON-NLS-1$
	/** The column visibility command. */
	public static final String		CMD_TOGGLE_COLUMN_VISIBILITY		= "Outline.ToggleColumnVisibility";	//$NON-NLS-1$
	private static final int		DIVIDER_HIT_SLOP					= 2;
// private static final int AUTO_SCROLL_MARGIN = 10;
	private OutlineModel			mModel;
	/** The header panel. */
	protected OutlineHeader			mHeaderPanel;
	private boolean					mDrawRowDividers;
	private boolean					mDrawColumnDividers;
	private Color					mDividerColor;
	private boolean					mDrawingDragImage;
	private Rectangle				mDragClip;
	private Column					mDividerDrag;
	private int						mColumnStart;
	private String					mSelectionChangedCommand;
	private String					mPotentialContentSizeChangeCommand;
	private boolean					mAllowColumnContextMenu;
	private boolean					mAllowColumnResize;
	private boolean					mAllowColumnDrag;
	private boolean					mAllowRowDrag;
	private String					mDefaultConfig;
	private boolean					mUseBanding;
	private ArrayList<Column>		mSavedColumns;
	private BufferedImage			mDownTriangle;
	private BufferedImage			mDownTriangleRoll;
	private BufferedImage			mRightTriangle;
	private BufferedImage			mRightTriangleRoll;
	private Row						mRollRow;
	private Row						mDragParentRow;
	private int						mDragChildInsertIndex;
	private boolean					mDragWasAcceptable;
	private boolean					mDragFocus;
	private Column					mSourceDragColumn;
	private boolean					mDynamicRowHeight;
	private HashSet<OutlineProxy>	mProxies;
	/** The first row index this outline will display. */
	protected int					mFirstRow;
	/** The last row index this outline will display. */
	protected int					mLastRow;
	private int						mSelectOnMouseUp;
	private RowFilter				mRowFilter;

	static {
		LocalizedMessages.initialize(Outline.class);
	}

	/** Creates a new outline. */
	public Outline() {
		this(true);
	}

	/**
	 * Creates a new outline.
	 * 
	 * @param model The model to use.
	 */
	public Outline(OutlineModel model) {
		this(model, true);
	}

	/**
	 * Creates a new outline.
	 * 
	 * @param showIndent Pass in <code>true</code> if the outline should show hierarchy and
	 *            controls for it.
	 */
	public Outline(boolean showIndent) {
		this(new OutlineModel(), showIndent);
	}

	/**
	 * Creates a new outline.
	 * 
	 * @param model The model to use.
	 * @param showIndent Pass in <code>true</code> if the outline should show hierarchy and
	 *            controls for it.
	 */
	public Outline(OutlineModel model, boolean showIndent) {
		super();
		mModel = model;
		mProxies = new HashSet<OutlineProxy>();
		mAllowColumnContextMenu = true;
		mAllowColumnResize = true;
		mAllowColumnDrag = true;
		mAllowRowDrag = true;
		mDrawRowDividers = true;
		mDrawColumnDividers = true;
		mUseBanding = true;
		mDividerColor = Color.LIGHT_GRAY;
		mSelectionChangedCommand = CMD_SELECTION_CHANGED;
		mPotentialContentSizeChangeCommand = CMD_POTENTIAL_CONTENT_SIZE_CHANGE;
		mDownTriangle = Images.getDownTriangleIcon();
		mDownTriangleRoll = Images.getDownTriangleRollIcon();
		mRightTriangle = Images.getRightTriangleIcon();
		mRightTriangleRoll = Images.getRightTriangleRollIcon();
		mDragChildInsertIndex = -1;
		mLastRow = -1;
		mModel.setShowIndent(showIndent);
		mModel.setIndentWidth(mDownTriangle.getWidth());

		setActionCommand(CMD_OPEN_SELECTION);
		setBackground(Color.white);
		setOpaque(true);
		setFocusable(true);
		addFocusListener(this);
		addMouseListener(this);
		addMouseMotionListener(this);
		addKeyListener(this);
		addComponentListener(this);
		setAutoscrolls(true);
		ToolTipManager.sharedInstance().registerComponent(this);

		if (!GraphicsUtilities.inHeadlessPrintMode()) {
			DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_COPY_OR_MOVE, this);
			setDropTarget(new DropTarget(this, this));
		}

		if (!(this instanceof OutlineProxy)) {
			mModel.addListener(this);
		}
	}

	/** @return The {@link RowFilter} being used. */
	public RowFilter getRowFilter() {
		return mRowFilter;
	}

	/** @param filter The {@link RowFilter} to use. */
	public void setRowFilter(RowFilter filter) {
		mRowFilter = filter;
	}

	/** Causes the {@link RowFilter} to be re-applied. */
	public void reapplyRowFilter() {
		ArrayList<Row> list = new ArrayList<Row>();
		for (Row row : mModel.getSelectionAsList()) {
			if (isRowFiltered(row)) {
				list.add(row);
			}
		}
		if (!list.isEmpty()) {
			mModel.deselect(list);
		}
		contentSizeMayHaveChanged();
		revalidateView();
	}

	private boolean isRowFiltered(Row row) {
		if (mRowFilter != null) {
			return mRowFilter.isRowFiltered(row);
		}
		return false;
	}

	/** @return The underlying data model. */
	public OutlineModel getModel() {
		return mModel;
	}

	/** @return This outline. */
	public Outline getRealOutline() {
		return this;
	}

	/** @param proxy The proxy to add. */
	protected void addProxy(OutlineProxy proxy) {
		mProxies.add(proxy);
		mModel.addListener(proxy);
	}

	/** Removes all proxies from this outline. */
	public void clearProxies() {
		for (OutlineProxy proxy : mProxies) {
			mModel.removeListener(proxy);
		}
		mProxies.clear();
	}

	/** @return Whether rows will resize vertically when their content changes. */
	public boolean dynamicRowHeight() {
		return mDynamicRowHeight;
	}

	/** @param dynamic Sets whether rows will resize vertically when their content changes. */
	public void setDynamicRowHeight(boolean dynamic) {
		mDynamicRowHeight = dynamic;
	}

	/** @return <code>true</code> if hierarchy indention (and controls) will be shown. */
	public boolean showIndent() {
		return mModel.showIndent();
	}

	/** @return The color to use when drawing the divider lines. */
	public Color getDividerColor() {
		return mDividerColor;
	}

	/** @param color The color to use when drawing the divider lines. */
	public void setDividerColor(Color color) {
		mDividerColor = color;
	}

	/** @return Whether to draw the row dividers or not. */
	public boolean shouldDrawRowDividers() {
		return mDrawRowDividers;
	}

	/** @param draw Whether to draw the row dividers or not. */
	public void setDrawRowDividers(boolean draw) {
		mDrawRowDividers = draw;
	}

	/** @return Whether to draw the column dividers or not. */
	public boolean shouldDrawColumnDividers() {
		return mDrawColumnDividers;
	}

	/** @param draw Whether to draw the column dividers or not. */
	public void setDrawColumnDividers(boolean draw) {
		mDrawColumnDividers = draw;
	}

	@Override public Dimension getPreferredSize() {
		Insets insets = getInsets();
		Dimension size = new Dimension(insets.left + insets.right, insets.top + insets.bottom);
		List<Column> columns = mModel.getColumns();
		boolean needRevalidate = false;

		for (Column col : columns) {
			int width = col.getWidth();
			if (width == -1) {
				width = col.getPreferredWidth(this);
				col.setWidth(width);
				needRevalidate = true;
			}
			if (col.isVisible()) {
				size.width += width + (mDrawColumnDividers ? 1 : 0);
			}
		}
		if (mDrawColumnDividers && !mModel.getColumns().isEmpty()) {
			size.width--;
		}

		if (needRevalidate) {
			revalidateView();
		}

		boolean needHeightAdjust = false;
		for (int i = getFirstRowToDisplay(); i <= getLastRowToDisplay(); i++) {
			Row row = mModel.getRowAtIndex(i);
			if (!isRowFiltered(row)) {
				int height = row.getHeight();
				if (height == -1) {
					height = row.getPreferredHeight(columns);
					row.setHeight(height);
				}
				size.height += height + (mDrawRowDividers ? 1 : 0);
				needHeightAdjust = true;
			}
		}
		if (mDrawRowDividers && needHeightAdjust) {
			size.height--;
		}

		if (isMinimumSizeSet()) {
			Dimension minSize = getMinimumSize();
			if (size.width < minSize.width) {
				size.width = minSize.width;
			}
			if (size.height < minSize.height) {
				size.height = minSize.height;
			}
		}
		return size;
	}

	private void drawDragRowInsertionMarker(Graphics gc, Row parent, int insertAtIndex) {
		Rectangle bounds = getDragRowInsertionMarkerBounds(parent, insertAtIndex);
		gc.setColor(Color.red);
		gc.drawLine(bounds.x, bounds.y + bounds.height / 2, bounds.x + bounds.width, bounds.y + bounds.height / 2);
		for (int i = 0; i < bounds.height / 2; i++) {
			gc.drawLine(bounds.x + i, bounds.y + i, bounds.x + i, bounds.y + bounds.height - (1 + i));
		}
	}

	private Rectangle getDragRowInsertionMarkerBounds(Row parent, int insertAtIndex) {
		int rowCount = mModel.getRowCount();
		Rectangle bounds;

		if (insertAtIndex < 0 || rowCount == 0) {
			bounds = new Rectangle();
		} else {
			int insertAt = getAbsoluteInsertionIndex(parent, insertAtIndex);
			int indent = parent != null ? mModel.getIndentWidth(parent, mModel.getColumns().get(0)) + mModel.getIndentWidth() : 0;
			if (insertAt < rowCount) {
				bounds = getRowBounds(mModel.getRowAtIndex(insertAt));
				if (mDrawRowDividers && insertAt != 0) {
					bounds.y--;
				}
			} else {
				bounds = getRowBounds(mModel.getRowAtIndex(rowCount - 1));
				bounds.y += bounds.height;
			}
			bounds.x += indent;
			bounds.width -= indent;
		}
		bounds.y -= 3;
		bounds.height = 7;
		return bounds;
	}

	/** @return The first row to display. By default, this would be 0. */
	public int getFirstRowToDisplay() {
		return mFirstRow < 0 ? 0 : mFirstRow;
	}

	/** @param index The first row to display. */
	public void setFirstRowToDisplay(int index) {
		mFirstRow = index;
	}

	/** @return The last row to display. */
	public int getLastRowToDisplay() {
		int max = mModel.getRowCount() - 1;
		return mLastRow < 0 || mLastRow > max ? max : mLastRow;
	}

	/**
	 * @param index The last row to display. If set to a negative value, then the last row index in
	 *            the outline will be returned from {@link #getLastRowToDisplay()}.
	 */
	public void setLastRowToDisplay(int index) {
		mLastRow = index;
	}

	@Override protected void paintComponent(Graphics gc) {
		super.paintComponent(GraphicsUtilities.prepare(gc));
		drawBackground(gc);

		Shape origClip = gc.getClip();
		Rectangle clip = gc.getClipBounds();
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		boolean active = isFocusOwner();
		int first = getFirstRowToDisplay();
		int last = getLastRowToDisplay();
		Container top = getTopLevelAncestor();
		boolean isPrinting = top instanceof AppWindow && ((AppWindow) top).isPrinting();
		boolean showIndent = showIndent();

		for (int rowIndex = first; rowIndex <= last; rowIndex++) {
			Row row = mModel.getRowAtIndex(rowIndex);
			if (!isRowFiltered(row)) {
				bounds.height = row.getHeight();
				if (bounds.y >= clip.y || bounds.y + bounds.height + (mDrawRowDividers ? 1 : 0) >= clip.y) {
					if (bounds.y > clip.y + clip.height) {
						break;
					}

					boolean rowSelected = !isPrinting && mModel.isRowSelected(row);
					if (!mDrawingDragImage || mDrawingDragImage && rowSelected) {
						Rectangle colBounds = new Rectangle(bounds);
						Composite savedComposite = null;
						boolean isFirstCol = true;
						int shift = 0;

						for (Column col : mModel.getColumns()) {
							if (col.isVisible()) {
								colBounds.width = col.getWidth();
								if (clip.intersects(colBounds)) {
									boolean dragging = mSourceDragColumn == col;
									gc.clipRect(colBounds.x, colBounds.y, colBounds.width, colBounds.height);
									if (dragging) {
										savedComposite = ((Graphics2D) gc).getComposite();
										((Graphics2D) gc).setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
									}
									if (showIndent && isFirstCol) {
										shift = mModel.getIndentWidth(row, col);
										colBounds.x += shift;
										colBounds.width -= shift;
										if (row.canHaveChildren()) {
											BufferedImage image = getDisclosureControl(row);
											gc.drawImage(image, colBounds.x - image.getWidth(), 1 + colBounds.y + (colBounds.height - image.getHeight()) / 2, null);
										}
									}
									// Under some circumstances, the width calculations
									// for cells are off by one pixel when printing...
									// so far, the only way I've found to compensate is
									// to put this hack in.
									if (isPrinting) {
										colBounds.width++;
									}
									col.drawRowCell(this, gc, colBounds, row, rowSelected, active);
									if (isPrinting) {
										colBounds.width--;
									}
									if (showIndent && isFirstCol) {
										colBounds.x -= shift;
										colBounds.width += shift;
									}
									if (dragging) {
										((Graphics2D) gc).setComposite(savedComposite);
									}
									gc.setClip(origClip);
								}
								colBounds.x += colBounds.width + 1;
								isFirstCol = false;
							}
						}

						if (mDrawingDragImage) {
							if (mDragClip == null) {
								mDragClip = new Rectangle(bounds);
							} else {
								mDragClip.add(bounds);
							}
						}
					}
				}
				bounds.y += bounds.height + (mDrawRowDividers ? 1 : 0);
			}
		}

		if (mDragChildInsertIndex != -1) {
			drawDragRowInsertionMarker(gc, mDragParentRow, mDragChildInsertIndex);
		}
	}

	private void drawBackground(Graphics gc) {
		super.paintComponent(gc);

		Rectangle clip = gc.getClipBounds();
		Insets insets = getInsets();
		int top = insets.top;
		int bottom = getHeight() - (top + insets.bottom);
		Rectangle bounds = new Rectangle(insets.left, top, getWidth() - (insets.left + insets.right), bottom);
		boolean active = isFocusOwner();
		int first = getFirstRowToDisplay();
		int last = getLastRowToDisplay();
		Container topAncestor = getTopLevelAncestor();
		boolean isPrinting = topAncestor instanceof AppWindow && ((AppWindow) topAncestor).isPrinting();

		for (int rowIndex = first; rowIndex <= last; rowIndex++) {
			Row row = mModel.getRowAtIndex(rowIndex);
			if (!isRowFiltered(row)) {
				bounds.height = row.getHeight();
				if (bounds.y >= clip.y || bounds.y + bounds.height + (mDrawRowDividers ? 1 : 0) >= clip.y) {
					if (bounds.y > clip.y + clip.height) {
						break;
					}
					boolean rowSelected = !isPrinting && mModel.isRowSelected(row);
					if (!mDrawingDragImage || mDrawingDragImage && rowSelected) {
						gc.setColor(getBackground(rowIndex, rowSelected, active));
						gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
						if (mDrawRowDividers) {
							gc.setColor(mDividerColor);
							gc.drawLine(bounds.x, bounds.y + bounds.height, bounds.x + bounds.width, bounds.y + bounds.height);
						}
					}
				}
				bounds.y += bounds.height + (mDrawRowDividers ? 1 : 0);
			}
		}

		if (mDrawColumnDividers) {
			int x = insets.left;
			bottom += top;
			gc.setColor(mDividerColor);
			for (Column col : mModel.getColumns()) {
				if (col.isVisible()) {
					x += col.getWidth();
					gc.drawLine(x, top, x, bottom);
					x++;
				}
			}
		}

		if (!isPrinting && (active || mDragFocus)) {
			gc.setColor(getListBackground(true, true));
			bounds = getVisibleRect();
			int x = bounds.x;
			int y = bounds.y;
			int w = bounds.width - 1;
			int h = bounds.height - 1;
			gc.drawRect(x, y, w, h);
		}
	}

	@Override public void repaint(Rectangle bounds) {
		super.repaint(bounds);
		// We have to check for null here, since repaint() will be called during
		// initialization of our super class.
		if (mProxies != null) {
			for (OutlineProxy proxy : mProxies) {
				proxy.repaintProxy(bounds);
			}
		}
	}

	/**
	 * Repaints the header panel, along with any proxy header panels.
	 * 
	 * @param bounds The bounds to repaint.
	 */
	void repaintHeader(Rectangle bounds) {
		if (mHeaderPanel != null) {
			getHeaderPanel().repaintInternal(bounds);
			for (OutlineProxy proxy : mProxies) {
				if (proxy.mHeaderPanel != null) {
					proxy.getHeaderPanel().repaintInternal(bounds);
				}
			}
		}
	}

	/** Repaints the header panel, if present. */
	void repaintHeader() {
		if (mHeaderPanel != null) {
			Rectangle bounds = mHeaderPanel.getBounds();
			bounds.x = 0;
			bounds.y = 0;
			mHeaderPanel.repaintInternal(bounds);
		}
	}

	/**
	 * Repaints the specified column index.
	 * 
	 * @param columnIndex The index of the column to repaint.
	 */
	public void repaintColumn(int columnIndex) {
		repaintColumn(mModel.getColumnAtIndex(columnIndex));
	}

	/**
	 * Repaints the specified column.
	 * 
	 * @param column The column to repaint.
	 */
	public void repaintColumn(Column column) {
		if (column.isVisible()) {
			Rectangle bounds = new Rectangle(getColumnStart(column), 0, column.getWidth(), getHeight());
			repaint(bounds);
			if (mHeaderPanel != null) {
				bounds.height = mHeaderPanel.getHeight();
				mHeaderPanel.repaint(bounds);
			}
		}
	}

	/** Repaints both the outline and its header. */
	public void repaintView() {
		repaint();
		if (mHeaderPanel != null) {
			mHeaderPanel.repaint();
		}
	}

	/** Repaints the current selection. */
	public void repaintSelection() {
		repaintSelectionInternal();
		for (OutlineProxy proxy : mProxies) {
			proxy.repaintSelectionInternal();
		}
	}

	/**
	 * Repaints the current selection.
	 * 
	 * @return The bounding rectangle of the repainted selection.
	 */
	protected Rectangle repaintSelectionInternal() {
		Rectangle area = new Rectangle();
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		int last = getLastRowToDisplay();
		List<Column> columns = mModel.getColumns();
		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			Row row = mModel.getRowAtIndex(i);
			if (!isRowFiltered(row)) {
				int height = row.getHeight();
				if (height == -1) {
					height = row.getPreferredHeight(columns);
					row.setHeight(height);
				}
				if (mDrawRowDividers) {
					height++;
				}
				if (mModel.isRowSelected(row)) {
					bounds.height = height;
					repaint(bounds);
					area = Geometry.union(area, bounds);
				}
				bounds.y += height;
			}
		}
		return area;
	}

	/**
	 * @param rowIndex The index of the row.
	 * @param selected Whether the row should be considered "selected".
	 * @param active Whether the outline should be considered "active".
	 * @return The background color for the specified row index.
	 */
	public Color getBackground(int rowIndex, boolean selected, boolean active) {
		if (selected) {
			return getListBackground(selected, active);
		}
		return useBanding() ? Colors.getBanding(rowIndex % 2 == 0) : Color.white;
	}

	public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
		return orientation == SwingConstants.VERTICAL ? visibleRect.height : visibleRect.width;
	}

	public Dimension getPreferredScrollableViewportSize() {
		return getPreferredSize();
	}

	public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
		if (orientation == SwingConstants.VERTICAL) {
			Insets insets = getInsets();
			int y = visibleRect.y - insets.top;
			int rowIndex;
			int rowTop;

			if (direction < 0) {
				rowIndex = overRowIndex(y);
				if (rowIndex > -1) {
					rowTop = getRowIndexStart(rowIndex);
					if (rowTop < y) {
						return y - rowTop;
					} else if (--rowIndex > -1) {
						return y - getRowIndexStart(rowIndex);
					}
				}
			} else {
				y += visibleRect.height;
				rowIndex = overRowIndex(y);
				if (rowIndex > -1) {
					rowTop = getRowIndexStart(rowIndex);
					int rowBottom = rowTop + mModel.getRowAtIndex(rowIndex).getHeight() + (mDrawRowDividers ? 1 : 0);
					if (rowBottom > y) {
						return rowBottom - (y - 1);
					} else if (++rowIndex < mModel.getRowCount()) {
						return getRowIndexStart(rowIndex) + mModel.getRowAtIndex(rowIndex).getHeight() + (mDrawRowDividers ? 1 : 0) - (y - 1);
					}
				}
			}
		}
		return 10;
	}

	public boolean getScrollableTracksViewportHeight() {
		return UIUtilities.shouldTrackViewportHeight(this);
	}

	public boolean getScrollableTracksViewportWidth() {
		return UIUtilities.shouldTrackViewportWidth(this);
	}

	/**
	 * Determines if the specified x-coordinate is over a column's divider.
	 * 
	 * @param x The coordinate to check.
	 * @return The column, or <code>null</code> if none is found.
	 */
	public Column overColumnDivider(int x) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();
		for (int i = 0; i < count; i++) {
			Column col = mModel.getColumnAtIndex(i);
			if (col.isVisible()) {
				pos += col.getWidth() + (mDrawColumnDividers ? 1 : 0);
				if (x >= pos - DIVIDER_HIT_SLOP && x <= pos + DIVIDER_HIT_SLOP) {
					return col;
				}
			}
		}
		return null;
	}

	/**
	 * Determines if the specified x-coordinate is over a column's divider.
	 * 
	 * @param x The coordinate to check.
	 * @return The column index, or <code>-1</code> if none is found.
	 */
	public int overColumnDividerIndex(int x) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();
		for (int i = 0; i < count; i++) {
			Column col = mModel.getColumnAtIndex(i);
			if (col.isVisible()) {
				pos += col.getWidth() + (mDrawColumnDividers ? 1 : 0);
				if (x >= pos - DIVIDER_HIT_SLOP && x <= pos + DIVIDER_HIT_SLOP) {
					return i;
				}
			}
		}
		return -1;
	}

	/**
	 * Determines if the specified x-coordinate is over a column.
	 * 
	 * @param x The coordinate to check.
	 * @return The column, or <code>null</code> if none is found.
	 */
	public Column overColumn(int x) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();
		for (int i = 0; i < count; i++) {
			Column col = mModel.getColumnAtIndex(i);
			if (col.isVisible()) {
				pos += col.getWidth() + (mDrawColumnDividers ? 1 : 0);
				if (x < pos) {
					return col;
				}
			}
		}
		return null;
	}

	/**
	 * Determines if the specified x-coordinate is over a column.
	 * 
	 * @param x The coordinate to check.
	 * @return The column index, or <code>-1</code> if none is found.
	 */
	public int overColumnIndex(int x) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();
		for (int i = 0; i < count; i++) {
			Column col = mModel.getColumnAtIndex(i);
			if (col.isVisible()) {
				pos += col.getWidth() + (mDrawColumnDividers ? 1 : 0);
				if (x < pos) {
					return i;
				}
			}
		}
		return -1;
	}

	private BufferedImage getDisclosureControl(Row row) {
		return row.isOpen() ? row == mRollRow ? mDownTriangleRoll : mDownTriangle : row == mRollRow ? mRightTriangleRoll : mRightTriangle;
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @param column The column the coordinates are currently over.
	 * @param row The row the coordinates are currently over.
	 * @return <code>true</code> if the coordinates are over a disclosure triangle.
	 */
	public boolean overDisclosureControl(int x, int y, Column column, Row row) {
		if (showIndent() && column != null && row != null && row.canHaveChildren() && mModel.isFirstColumn(column)) {
			BufferedImage image = getDisclosureControl(row);
			int right = getInsets().left + mModel.getIndentWidth(row, column);
			return x <= right && x >= right - image.getWidth();
		}
		return false;
	}

	/**
	 * @param columnIndex The index of the column.
	 * @return The starting x-coordinate for the specified column index.
	 */
	public int getColumnIndexStart(int columnIndex) {
		int pos = getInsets().left;
		for (int i = 0; i < columnIndex; i++) {
			Column column = mModel.getColumnAtIndex(i);
			if (column.isVisible()) {
				pos += column.getWidth() + (mDrawColumnDividers ? 1 : 0);
			}
		}
		return pos;
	}

	/**
	 * @param column The column.
	 * @return The starting x-coordinate for the specified column.
	 */
	public int getColumnStart(Column column) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();
		for (int i = 0; i < count; i++) {
			Column col = mModel.getColumnAtIndex(i);
			if (col == column) {
				break;
			}
			if (col.isVisible()) {
				pos += col.getWidth() + (mDrawColumnDividers ? 1 : 0);
			}
		}
		return pos;
	}

	/**
	 * @param column The column.
	 * @return An {@link BufferedImage}containing the drag image for the specified column.
	 */
	public BufferedImage getColumnDragImage(Column column) {
		BufferedImage offscreen = null;
		synchronized (getTreeLock()) {
			Graphics2D g2d = null;
			try {
				Rectangle bounds = new Rectangle(0, 0, column.getWidth() + (mDrawColumnDividers ? 2 : 0), getVisibleRect().height + (mHeaderPanel != null ? mHeaderPanel.getHeight() + 1 : 0));
				offscreen = getGraphicsConfiguration().createCompatibleImage(bounds.width, bounds.height, Transparency.TRANSLUCENT);
				g2d = GraphicsUtilities.prepare(offscreen.getGraphics());
				g2d.setClip(bounds);
				g2d.setBackground(new Color(0, true));
				g2d.clearRect(0, 0, bounds.width, bounds.height);
				g2d.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
				g2d.setColor(getBackground());
				g2d.fill(bounds);
				g2d.setColor(getDividerColor());
				if (mDrawRowDividers) {
					g2d.drawLine(bounds.x, bounds.y, bounds.x + bounds.width, bounds.y);
					g2d.drawLine(bounds.x, bounds.y + bounds.height - 1, bounds.x + bounds.width, bounds.y + bounds.height - 1);
					bounds.y++;
					bounds.height -= 2;
				}
				if (mDrawColumnDividers) {
					g2d.drawLine(bounds.x, bounds.y, bounds.x, bounds.y + bounds.height);
					g2d.drawLine(bounds.x + bounds.width - 1, bounds.y, bounds.x + bounds.width - 1, bounds.y + bounds.height);
					bounds.x++;
					bounds.width -= 2;
				}
				drawOneColumn(g2d, column, bounds);
			} catch (Exception exception) {
				assert false : Debug.throwableToString(exception);
			} finally {
				if (g2d != null) {
					g2d.dispose();
				}
			}
		}
		return offscreen;
	}

	private void drawOneColumn(Graphics2D g2d, Column column, Rectangle bounds) {
		Shape oldClip = g2d.getClip();
		Color divColor = getDividerColor();
		int last = getLastRowToDisplay();
		int y = bounds.y;
		int maxY = bounds.y + bounds.height;

		if (mHeaderPanel != null) {
			bounds.height = mHeaderPanel.getHeight();
			g2d.setColor(mHeaderPanel.getBackground());
			g2d.fill(bounds);
			column.drawHeaderCell(this, g2d, bounds);
			bounds.y += mHeaderPanel.getHeight();
			g2d.setColor(divColor);
			g2d.drawLine(bounds.x, bounds.y, bounds.x + bounds.width, bounds.y);
			bounds.y++;
		}

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			Row row = mModel.getRowAtIndex(i);
			if (!isRowFiltered(row)) {
				bounds.height = row.getHeight();
				if (maxY < bounds.y) {
					break;
				}
				if (y <= bounds.y) {
					g2d.setClip(bounds);
					g2d.setColor(getBackground(i, false, true));
					g2d.fill(bounds);
					column.drawRowCell(this, g2d, bounds, row, false, true);
					g2d.setClip(oldClip);
					if (mDrawRowDividers) {
						g2d.setColor(divColor);
						g2d.drawLine(bounds.x, bounds.y + bounds.height, bounds.x + bounds.width, bounds.y + bounds.height);
					}
				}
				bounds.y += bounds.height + (mDrawRowDividers ? 1 : 0);
			}
		}
	}

	/**
	 * Determines if the specified y-coordinate is over a row.
	 * 
	 * @param y The coordinate to check.
	 * @return The row, or <code>null</code> if none is found.
	 */
	public Row overRow(int y) {
		List<Row> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();
		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			Row row = rows.get(i);
			if (!isRowFiltered(row)) {
				pos += row.getHeight() + (mDrawRowDividers ? 1 : 0);
				if (y < pos) {
					return row;
				}
			}
		}
		return null;
	}

	/**
	 * Determines if the specified y-coordinate is over a row.
	 * 
	 * @param y The coordinate to check.
	 * @return The row index, or <code>-1</code> if none is found.
	 */
	public int overRowIndex(int y) {
		List<Row> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();
		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			Row row = rows.get(i);
			if (!isRowFiltered(row)) {
				pos += row.getHeight() + (mDrawRowDividers ? 1 : 0);
				if (y < pos) {
					return i;
				}
			}
		}
		return -1;
	}

	/**
	 * @param y The coordinate to check.
	 * @return The row index to insert at, from <code>0</code> to
	 *         {@link OutlineModel#getRowCount()}.
	 */
	public int getRowInsertionIndex(int y) {
		List<Row> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();
		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			Row row = rows.get(i);
			if (!isRowFiltered(row)) {
				int height = row.getHeight();
				int tmp = pos + height / 2;
				if (y <= tmp) {
					return i;
				}
				pos += height + (mDrawRowDividers ? 1 : 0);
			}
		}
		return last;
	}

	/**
	 * @param index The index of the row.
	 * @return The starting y-coordinate for the specified row index.
	 */
	public int getRowIndexStart(int index) {
		List<Row> rows = mModel.getRows();
		int pos = getInsets().top;
		for (int i = getFirstRowToDisplay(); i < index; i++) {
			Row row = rows.get(i);
			if (!isRowFiltered(row)) {
				pos += row.getHeight() + (mDrawRowDividers ? 1 : 0);
			}
		}
		return pos;
	}

	/**
	 * @param row The row.
	 * @return The starting y-coordinate for the specified row.
	 */
	public int getRowStart(Row row) {
		List<Row> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();
		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			Row oneRow = rows.get(i);
			if (row == oneRow) {
				break;
			}
			if (!isRowFiltered(oneRow)) {
				pos += oneRow.getHeight() + (mDrawRowDividers ? 1 : 0);
			}
		}
		return pos;
	}

	/**
	 * @param rowIndex The index of the row.
	 * @return The bounds of the row at the specified index.
	 */
	public Rectangle getRowIndexBounds(int rowIndex) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		bounds.y = getRowIndexStart(rowIndex);
		bounds.height = mModel.getRowAtIndex(rowIndex).getHeight();
		return bounds;
	}

	/**
	 * @param row The row.
	 * @return The bounds of the specified row.
	 */
	public Rectangle getRowBounds(Row row) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		bounds.y = getRowStart(row);
		bounds.height = row.getHeight();
		return bounds;
	}

	/**
	 * @param row The row to use.
	 * @param column The column to use.
	 * @return The bounds of the specified cell.
	 */
	public Rectangle getCellBounds(Row row, Column column) {
		Rectangle bounds = getRowBounds(row);
		bounds.x = getColumnStart(column);
		bounds.width = column.getWidth();
		return bounds;
	}

	/**
	 * @param row The row to use.
	 * @param column The column to use.
	 * @return The bounds of the specified cell, adjusted for any necessary indent.
	 */
	public Rectangle getAdjustedCellBounds(Row row, Column column) {
		Rectangle bounds = getCellBounds(row, column);
		if (mModel.isFirstColumn(column)) {
			int indent = mModel.getIndentWidth(row, column);
			bounds.x += indent;
			bounds.width -= indent;
			if (bounds.width < 1) {
				bounds.width = 1;
			}
		}
		return bounds;
	}

	/** Sets the width of all visible columns to their preferred width. */
	public void sizeColumnsToFit() {
		ArrayList<Column> columns = new ArrayList<Column>();
		for (Column column : mModel.getColumns()) {
			if (column.isVisible()) {
				int width = column.getPreferredWidth(this);
				if (width != column.getWidth()) {
					column.setWidth(width);
					columns.add(column);
				}
			}
		}
		if (!columns.isEmpty()) {
			processColumnWidthChanges(columns);
		}
	}

	/**
	 * Repaint the specified row in this outline, as well as its proxies.
	 * 
	 * @param row The row to repaint.
	 */
	protected void repaintProxyRow(Row row) {
		repaint(getRowBounds(row));
		for (OutlineProxy proxy : mProxies) {
			proxy.repaintProxy(proxy.getRowBounds(row));
		}
	}

	/** Notifies all action listeners of a selection change. */
	protected void notifyOfSelectionChange() {
		notifyActionListeners(new ActionEvent(getRealOutline(), ActionEvent.ACTION_PERFORMED, getSelectionChangedActionCommand()));
	}

	/** @return The selection changed action command. */
	public String getSelectionChangedActionCommand() {
		return mSelectionChangedCommand;
	}

	/** @param command The selection changed action command. */
	public void setSelectionChangedActionCommand(String command) {
		mSelectionChangedCommand = command;
	}

	/** @return The potential user-initiated content size change action command. */
	public String getPotentialContentSizeChangeActionCommand() {
		return mPotentialContentSizeChangeCommand;
	}

	/** @param command The potential user-initiated content size change action command. */
	public void setPotentialContentSizeChangeActionCommand(String command) {
		mPotentialContentSizeChangeCommand = command;
	}

	/** @return Whether the selection can be deleted. */
	public boolean canDeleteSelection() {
		return false;
	}

	/** Delete the currently selected rows. */
	public void deleteSelection() {
		// Does nothing by default.
	}

	/**
	 * Arranges the columns in the same order as the columns passed in.
	 * 
	 * @param columns The column order.
	 */
	public void setColumnOrder(List<Column> columns) {
		ArrayList<Column> list = new ArrayList<Column>(columns);
		List<Column> cols = mModel.getColumns();
		cols.removeAll(columns);
		list.addAll(cols);
		cols.clear();
		cols.addAll(list);
		repaint();
		getHeaderPanel().repaint();
	}

	/** @return The header panel for this table. */
	public OutlineHeader getHeaderPanel() {
		if (mHeaderPanel == null) {
			mHeaderPanel = new OutlineHeader(this);
		}
		return mHeaderPanel;
	}

	/** @return The source column being dragged. */
	public Column getSourceDragColumn() {
		return mSourceDragColumn;
	}

	/** @param column The source column being dragged. */
	protected void setSourceDragColumn(Column column) {
		if (mSourceDragColumn != null) {
			repaintColumn(mSourceDragColumn);
		}
		mSourceDragColumn = column;
		if (mSourceDragColumn != null) {
			repaintColumn(mSourceDragColumn);
		}
	}

	/** @param scrollTo The row index to scroll to. */
	protected void keyScroll(int scrollTo) {
		Outline real = getRealOutline();
		if (!keyScrollInternal(real, scrollTo)) {
			for (OutlineProxy proxy : real.mProxies) {
				if (keyScrollInternal(proxy, scrollTo)) {
					break;
				}
			}
		}
	}

	private boolean keyScrollInternal(Outline outline, int scrollTo) {
		if (scrollTo >= outline.mFirstRow && scrollTo <= outline.mLastRow) {
			outline.requestFocus();
			outline.scrollRectToVisible(outline.getRowIndexBounds(scrollTo));
			return true;
		}
		return false;
	}

	public void mouseClicked(MouseEvent event) {
		Row rollRow = null;
		try {
			boolean local = event.getSource() == this;
			int x = event.getX();
			int y = event.getY();
			Row rowHit;
			Column column;
			int clickCount = event.getClickCount();

			rollRow = mRollRow;
			if (clickCount == 1) {
				if (local) {
					column = overColumn(x);
					if (column != null) {
						rowHit = overRow(y);
						if (rowHit != null && !overDisclosureControl(x, y, column, rowHit)) {
							Cell cell = column.getRowCell(rowHit);
							if (cell != null) {
								cell.mouseClicked(event, getCellBounds(rowHit, column), rowHit, column);
							}
						}
					}
				}
			} else if (clickCount == 2) {
				column = overColumnDivider(x);
				if (column == null) {
					if (local) {
						rowHit = overRow(y);
						if (rowHit != null) {
							column = overColumn(x);
							if ((column == null || !overDisclosureControl(x, y, column, rowHit)) && mModel.isRowSelected(rowHit)) {
								notifyActionListeners();
							}
						}
					}
				} else if (allowColumnResize()) {
					int width = column.getPreferredWidth(this);
					if (width != column.getWidth()) {
						adjustColumnWidth(column, width);
					}
				}
			}
		} finally {
			repaintChangedRollRow(rollRow);
		}
	}

	private void repaintChangedRollRow(Row rollRow) {
		if (rollRow != mRollRow) {
			Column column = mModel.getColumnAtIndex(0);
			Rectangle bounds;

			if (mRollRow != null) {
				bounds = getCellBounds(mRollRow, column);
				bounds.width = mModel.getIndentWidth(mRollRow, column);
				repaint(bounds);
			}
			if (rollRow != null) {
				bounds = getCellBounds(rollRow, column);
				bounds.width = mModel.getIndentWidth(rollRow, column);
				repaint(bounds);
			}
			mRollRow = rollRow;
		}
	}

	public void mouseEntered(MouseEvent event) {
		// Not used.
	}

	public void mouseExited(MouseEvent event) {
		// Not used.
	}

	public void mousePressed(MouseEvent event) {
		if (isEnabled()) {
			Row rollRow = null;
			try {
				boolean local = event.getSource() == this;
				int x = event.getX();
				int y = event.getY();
				Row rowHit;
				int rowIndexHit;
				Column column;
				requestFocus();
				mSelectOnMouseUp = -1;
				mDividerDrag = overColumnDivider(x);
				if (mDividerDrag != null) {
					if (allowColumnResize()) {
						mColumnStart = getColumnStart(mDividerDrag);
					}
				} else if (local) {
					column = overColumn(x);
					rowHit = overRow(y);
					if (column != null && rowHit != null) {
						if (overDisclosureControl(x, y, column, rowHit)) {
							Rectangle bounds = getCellBounds(rowHit, column);
							bounds.width = mModel.getIndentWidth(rowHit, column);
							rollRow = mRollRow;
							repaint(bounds);
							rowHit.setOpen(!rowHit.isOpen());
							return;
						}
					}
					int method = Selection.MOUSE_NONE;

					rowIndexHit = overRowIndex(y);

					if (event.isShiftDown()) {
						method |= Selection.MOUSE_EXTEND;
					}
					if ((event.getModifiers() & getToolkit().getMenuShortcutKeyMask()) != 0 && !event.isPopupTrigger()) {
						method |= Selection.MOUSE_FLIP;
					}
					mSelectOnMouseUp = mModel.getSelection().selectByMouse(rowIndexHit, method);
					if (event.isPopupTrigger()) {
						mSelectOnMouseUp = -1;
						showContextMenu(event);
					}
				}
			} finally {
				repaintChangedRollRow(rollRow);
			}
		}
	}

	public void mouseReleased(MouseEvent event) {
		if (isEnabled()) {
			Row rollRow = null;
			try {
				rollRow = mRollRow;
				if (mDividerDrag != null && allowColumnResize()) {
					dragColumnDivider(event.getX());
				}
				mDividerDrag = null;
				if (mSelectOnMouseUp != -1) {
					mModel.select(mSelectOnMouseUp, false);
					mSelectOnMouseUp = -1;
				}
			} finally {
				repaintChangedRollRow(rollRow);
			}
		}
	}

	public void mouseDragged(MouseEvent event) {
		if (isEnabled()) {
			Row rollRow = null;
			try {
				int x = event.getX();
				mSelectOnMouseUp = -1;
				if (mDividerDrag != null && allowColumnResize()) {
					dragColumnDivider(x);
					JScrollPane scrollPane = (JScrollPane) UIUtilities.getAncestorOfType(this, JScrollPane.class);
					if (scrollPane != null) {
						Point pt = event.getPoint();
						if (!(event.getSource() instanceof Outline)) {
							// Column resizing is occurring in the header, most likely
							pt.y = getVisibleRect().y + 1;
						}
						scrollRectToVisible(new Rectangle(pt.x, pt.y, 1, 1));
					}
				}
			} finally {
				repaintChangedRollRow(rollRow);
			}
		}
	}

	public void mouseMoved(MouseEvent event) {
		if (isEnabled()) {
			Row rollRow = null;
			try {
				boolean local = event.getSource() == this;
				int x = event.getX();
				int y = event.getY();
				Row rowHit;
				Column column;

				Cursor cursor = Cursor.getDefaultCursor();

				if (overColumnDivider(x) != null) {
					if (allowColumnResize()) {
						cursor = Cursor.getPredefinedCursor(Cursor.W_RESIZE_CURSOR);
					}
				} else if (local) {
					column = overColumn(x);
					if (column != null) {
						rowHit = overRow(y);
						if (rowHit != null) {
							if (overDisclosureControl(x, y, column, rowHit)) {
								rollRow = rowHit;
							} else {
								Cell cell = column.getRowCell(rowHit);
								cursor = cell.getCursor(event, getCellBounds(rowHit, column), rowHit, column);
							}
						}
					}
				}
				setCursor(cursor);
			} finally {
				repaintChangedRollRow(rollRow);
			}
		}
	}

	public void keyPressed(KeyEvent event) {
		Selection selection = mModel.getSelection();
		boolean shiftDown = event.isShiftDown();
		int index;
		switch (event.getKeyCode()) {
			case KeyEvent.VK_LEFT:
				index = selection.firstSelectedIndex();
				while (index >= 0) {
					mModel.getRowAtIndex(index).setOpen(false);
					index = selection.nextSelectedIndex(index + 1);
				}
				break;
			case KeyEvent.VK_RIGHT:
				index = selection.firstSelectedIndex();
				while (index >= 0) {
					mModel.getRowAtIndex(index).setOpen(true);
					index = selection.nextSelectedIndex(index + 1);
				}
				break;
			case KeyEvent.VK_UP:
				keyScroll(selection.selectUp(shiftDown));
				break;
			case KeyEvent.VK_DOWN:
				keyScroll(selection.selectDown(shiftDown));
				break;
			case KeyEvent.VK_HOME:
				keyScroll(selection.selectToHome(shiftDown));
				break;
			case KeyEvent.VK_END:
				keyScroll(selection.selectToEnd(shiftDown));
				break;
			default:
				return;
		}
		event.consume();
	}

	public void keyTyped(KeyEvent event) {
		char ch = event.getKeyChar();
		if (ch == '\n' || ch == '\r') {
			if (mModel.hasSelection()) {
				notifyActionListeners();
			}
			event.consume();
		} else if (ch == '\b' || ch == KeyEvent.VK_DELETE) {
			if (canDeleteSelection()) {
				deleteSelection();
			}
			event.consume();
		}
	}

	public void keyReleased(KeyEvent event) {
		// Not used.
	}

	/**
	 * @param viewPt The location within the view.
	 * @return The row cell at the specified point.
	 */
	public Cell getCellAt(Point viewPt) {
		return getCellAt(viewPt.x, viewPt.y);
	}

	/**
	 * @param x The x-coordinate within the view.
	 * @param y The y-coordinate within the view.
	 * @return The row cell at the specified coordinates.
	 */
	public Cell getCellAt(int x, int y) {
		Column column = overColumn(x);
		if (column != null) {
			Row row = overRow(y);
			if (row != null) {
				return column.getRowCell(row);
			}
		}
		return null;
	}

	private void dragColumnDivider(int x) {
		int old = mDividerDrag.getWidth();
		if (x <= mColumnStart + DIVIDER_HIT_SLOP * 2) {
			x = mColumnStart + DIVIDER_HIT_SLOP * 2 + 1;
		}
		x -= mColumnStart;
		if (old != x) {
			adjustColumnWidth(mDividerDrag, x);
		}
	}

	/**
	 * @param column The column to adjust.
	 * @param width The new column width.
	 */
	public void adjustColumnWidth(Column column, int width) {
		ArrayList<Column> columns = new ArrayList<Column>(1);
		column.setWidth(width);
		columns.add(column);
		processColumnWidthChanges(columns);
	}

	private void processColumnWidthChanges(ArrayList<Column> columns) {
		updateRowHeightsIfNeeded(columns);
		revalidateView();
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The drag image for this table when dragging rows.
	 */
	protected BufferedImage getDragImage(int x, int y) {
		Graphics2D g2d = null;
		BufferedImage off1 = null;
		BufferedImage off2 = null;

		mDrawingDragImage = true;
		mDragClip = null;
		off1 = getImage();
		mDrawingDragImage = false;

		if (mDragClip == null) {
			mDragClip = new Rectangle(x, y, 1, 1);
		}

		try {
			off2 = getGraphicsConfiguration().createCompatibleImage(mDragClip.width, mDragClip.height, Transparency.TRANSLUCENT);
			g2d = (Graphics2D) off2.getGraphics();
			g2d.setClip(new Rectangle(0, 0, mDragClip.width, mDragClip.height));
			g2d.setBackground(new Color(0, true));
			g2d.clearRect(0, 0, mDragClip.width, mDragClip.height);
			g2d.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
			Rectangle bounds = getVisibleRect();
			g2d.translate(-(mDragClip.x - bounds.x), -(mDragClip.y - bounds.y));
			g2d.drawImage(off1, 0, 0, this);
		} catch (Exception paintException) {
			assert false : Debug.throwableToString(paintException);
			off2 = null;
			mDragClip = new Rectangle(x, y, 1, 1);
		} finally {
			if (g2d != null) {
				g2d.dispose();
			}
		}

		return off2 != null ? off2 : off1;
	}

	/**
	 * @return An {@link BufferedImage} containing the current contents of this component, minus the
	 *         specified component and its children.
	 */
	public BufferedImage getImage() {
		BufferedImage offscreen = null;

		synchronized (getTreeLock()) {
			Graphics2D g2d = null;

			try {
				Rectangle bounds = getVisibleRect();
				offscreen = getGraphicsConfiguration().createCompatibleImage(bounds.width, bounds.height, Transparency.TRANSLUCENT);
				g2d = (Graphics2D) offscreen.getGraphics();
				Color saved = g2d.getBackground();
				g2d.setBackground(new Color(0, true));
				g2d.clearRect(0, 0, bounds.width, bounds.height);
				g2d.setBackground(saved);
				Rectangle clip = new Rectangle(0, 0, bounds.width, bounds.height);
				g2d.setClip(clip);
				g2d.translate(-bounds.x, -bounds.y);
				paint(g2d);
			} catch (Exception exception) {
				assert false : Debug.throwableToString(exception);
			} finally {
				if (g2d != null) {
					g2d.dispose();
				}
			}
		}
		return offscreen;
	}

	@Override public boolean isOpaque() {
		return super.isOpaque() && !mDrawingDragImage;
	}

	/**
	 * Displays a context menu.
	 * 
	 * @param event The triggering mouse event.
	 */
	protected void showContextMenu(MouseEvent event) {
		// RAW: Implement?
	}

	/** @return <code>true</code> if column resizing is allowed. */
	public boolean allowColumnResize() {
		return mAllowColumnResize;
	}

	/** @param allow Whether column resizing is on or off. */
	public void setAllowColumnResize(boolean allow) {
		mAllowColumnResize = allow;
	}

	/** @return <code>true</code> if column dragging is allowed. */
	public boolean allowColumnDrag() {
		return mAllowColumnDrag;
	}

	/** @param allow Whether column dragging is on or off. */
	public void setAllowColumnDrag(boolean allow) {
		mAllowColumnDrag = allow;
	}

	/** @return <code>true</code> if row dragging is allowed. */
	public boolean allowRowDrag() {
		return mAllowRowDrag;
	}

	/** @param allow Whether row dragging is on or off. */
	public void setAllowRowDrag(boolean allow) {
		mAllowRowDrag = allow;
	}

	/** @return <code>true</code> if the column context menu is allowed. */
	public boolean allowColumnContextMenu() {
		return mAllowColumnContextMenu;
	}

	/** @param allow Whether the column context menu is on or off. */
	public void setAllowColumnContextMenu(boolean allow) {
		mAllowColumnContextMenu = allow;
	}

	/** Revalidates the view and header panel if it exists. */
	public void revalidateView() {
		revalidate();
		if (mHeaderPanel != null) {
			mHeaderPanel.revalidate();
			mHeaderPanel.repaint();
		}
		repaint();
	}

	@Override public void setBounds(int x, int y, int width, int height) {
		super.setBounds(x, y, width, height);
		if (mHeaderPanel != null) {
			Dimension size = mHeaderPanel.getPreferredSize();
			mHeaderPanel.setResizeOK(true);
			mHeaderPanel.setBounds(getX(), mHeaderPanel.getY(), getWidth(), size.height);
			mHeaderPanel.setResizeOK(false);
		}
	}

	/**
	 * Sort on a column.
	 * 
	 * @param column The column to sort on.
	 * @param ascending Pass in <code>true</code> for an ascending sort.
	 * @param add Pass in <code>true</code> to add this column to the end of the sort order, or
	 *            <code>false</code> to make this column the primary and only sort column.
	 */
	public void setSort(Column column, boolean ascending, boolean add) {
		StateEdit edit = new StateEdit(mModel, MSG_SORT_UNDO_TITLE);
		int count = mModel.getColumnCount();
		int i;

		if (!add) {
			for (i = 0; i < count; i++) {
				Column col = mModel.getColumnAtIndex(i);
				if (column == col) {
					col.setSortCriteria(0, ascending);
				} else {
					col.setSortCriteria(-1, col.isSortAscending());
				}
			}
		} else {
			if (column.getSortSequence() == -1) {
				int highest = -1;
				for (i = 0; i < count; i++) {
					int sortOrder = mModel.getColumnAtIndex(i).getSortSequence();
					if (sortOrder > highest) {
						highest = sortOrder;
					}
				}
				column.setSortCriteria(highest + 1, ascending);
			} else {
				column.setSortCriteria(column.getSortSequence(), ascending);
			}
		}
		mModel.sort();
		edit.end();
		postUndo(edit);
	}

	public void focusGained(FocusEvent event) {
		repaintSelection();
		repaintFocus();
	}

	public void focusLost(FocusEvent event) {
		repaintSelection();
		repaintFocus();
	}

	private void repaintFocus() {
		Rectangle bounds = getVisibleRect();
		int x = bounds.x;
		int y = bounds.y;
		int w = bounds.width;
		int h = bounds.height;
		paintImmediately(x, y, w, 1);
		paintImmediately(x, y, 1, h);
		paintImmediately(x + w - 1, y, 1, h);
		paintImmediately(x, y + h - 1, w, 1);
	}

	/**
	 * @param outline The outline to check.
	 * @return Whether the specified outline refers to this outline or a proxy of it.
	 */
	public boolean isSelfOrProxy(Outline outline) {
		Outline self = getRealOutline();
		if (outline == self) {
			return true;
		}
		for (OutlineProxy proxy : self.mProxies) {
			if (outline == proxy) {
				return true;
			}
		}
		return false;
	}

	@Override public Point getToolTipLocation(MouseEvent event) {
		int x = event.getX();
		Column column = overColumn(x);
		if (column != null) {
			Row row = overRow(event.getY());
			if (row != null) {
				Rectangle bounds = getCellBounds(row, column);
				return new Point(x, bounds.y + bounds.height);
			}
		}
		return null;
	}

	@Override public String getToolTipText(MouseEvent event) {
		Column column = overColumn(event.getX());
		if (column != null) {
			Row row = overRow(event.getY());
			if (row != null) {
				return column.getRowCell(row).getToolTipText(event, getCellBounds(row, column), row, column);
			}
		}
		return super.getToolTipText(event);
	}

	/** @return <code>true</code> if background banding is enabled. */
	public boolean useBanding() {
		return mUseBanding;
	}

	/** @param useBanding Whether to use background banding or not. */
	public void setUseBanding(boolean useBanding) {
		mUseBanding = useBanding;
	}

	/**
	 * Creates a configuration that can be applied to an outline.
	 * 
	 * @param configSpec The column configuration spec for each column.
	 * @return The configuration.
	 */
	public static String createConfig(ColumnConfig[] configSpec) {
		return createConfig(configSpec, 0, 0);
	}

	/**
	 * Creates a configuration that can be applied to an outline.
	 * 
	 * @param configSpec The column configuration spec for each column.
	 * @param hSplit The position of the horizontal splitter.
	 * @param vSplit The position of the vertical splitter.
	 * @return The configuration.
	 */
	public static String createConfig(ColumnConfig[] configSpec, int hSplit, int vSplit) {
		StringBuilder buffer = new StringBuilder();
		buffer.append(OutlineModel.CONFIG_VERSION);
		buffer.append('\t');
		buffer.append(configSpec.length);
		for (ColumnConfig element : configSpec) {
			buffer.append('\t');
			buffer.append(element.mID);
			buffer.append('\t');
			buffer.append(element.mVisible);
			buffer.append('\t');
			buffer.append(element.mWidth);
			buffer.append('\t');
			buffer.append(element.mSortSequence);
			buffer.append('\t');
			buffer.append(element.mSortAscending);
		}
		buffer.append('\t');
		buffer.append(hSplit);
		buffer.append('\t');
		buffer.append(vSplit);
		return buffer.toString();
	}

	/**
	 * @return A configuration string that can be used to restore the current column configuration
	 *         and splitter settings (if the outline is embedded in a scroll panel).
	 */
	public String getConfig() {
		StringBuilder buffer = new StringBuilder();
		int count = mModel.getColumnCount();
		buffer.append(OutlineModel.CONFIG_VERSION);
		buffer.append('\t');
		buffer.append(count);
		for (int i = 0; i < count; i++) {
			Column column = mModel.getColumnAtIndex(i);
			buffer.append('\t');
			buffer.append(column.getID());
			buffer.append('\t');
			buffer.append(column.isVisible());
			buffer.append('\t');
			buffer.append(column.getWidth());
			buffer.append('\t');
			buffer.append(column.getSortSequence());
			buffer.append('\t');
			buffer.append(column.isSortAscending());
		}
		return buffer.toString();
	}

	private int getInteger(StringTokenizer tokenizer, int def) {
		try {
			return Integer.parseInt(tokenizer.nextToken().trim());
		} catch (Exception exception) {
			return def;
		}
	}

	/**
	 * Attempts to restore the specified column configuration.
	 * 
	 * @param config The configuration to restore.
	 */
	public void applyConfig(String config) {
		try {
			StringTokenizer tokenizer = new StringTokenizer(config, "\t"); //$NON-NLS-1$
			if (getInteger(tokenizer, 0) == OutlineModel.CONFIG_VERSION) {
				int count = getInteger(tokenizer, 0);
				List<Column> columns = mModel.getColumns();
				boolean needSort = false;
				Column column;
				int i;

				mModel.clearSort();

				for (i = 0; i < count; i++) {
					column = mModel.getColumnWithID(getInteger(tokenizer, 0));
					if (column == null) {
						throw new Exception();
					}
					columns.remove(column);
					columns.add(i, column);
					column.setVisible(NumberUtils.getBoolean(tokenizer.nextToken()));
					column.setWidth(getInteger(tokenizer, column.getWidth()));
					column.setSortCriteria(getInteger(tokenizer, -1), NumberUtils.getBoolean(tokenizer.nextToken()));
					if (column.getSortSequence() != -1) {
						needSort = true;
					}
				}
				if (needSort) {
					mModel.sort();
				}
				updateRowHeightsIfNeeded(columns);
			}
		} catch (Exception exception) {
			// Nothing can be done, so allow the view to restore itself
		}

		revalidateView();
	}

	/** @return The default configuration. */
	public String getDefaultConfig() {
		if (mDefaultConfig == null) {
			mDefaultConfig = getConfig();
		}
		return mDefaultConfig;
	}

	/** @param config The configuration to set as the default. */
	public void setDefaultConfig(String config) {
		mDefaultConfig = config;
	}

	public void dragGestureRecognized(DragGestureEvent dge) {
		DragUtil.prepDrag();
		if (mDividerDrag == null && mModel.hasSelection() && allowRowDrag()) {
			Point pt = dge.getDragOrigin();
			RowSelection selection = new RowSelection(mModel, mModel.getSelectionAsList(true).toArray(new Row[0]));
			if (DragSource.isDragImageSupported()) {
				BufferedImage dragImage = getDragImage(pt.x, pt.y);
				Point imageOffset = new Point(mDragClip.x - pt.x, mDragClip.y - pt.y);
				dge.startDrag(null, dragImage, imageOffset, selection, null);
			} else {
				dge.startDrag(null, selection);
			}
		}
	}

	/**
	 * @param dtde The drop target drag event.
	 * @return <code>true</code> if the contents of the drag can be dropped into this outline.
	 */
	protected boolean isDragAcceptable(DropTargetDragEvent dtde) {
		boolean result = false;
		try {
			if (dtde.isDataFlavorSupported(Column.DATA_FLAVOR)) {
				Column column = (Column) dtde.getTransferable().getTransferData(Column.DATA_FLAVOR);
				result = isColumnDragAcceptable(dtde, column);
				if (result) {
					mModel.setDragColumn(column);
				}
			}
			if (dtde.isDataFlavorSupported(RowSelection.DATA_FLAVOR)) {
				Row[] rows = (Row[]) dtde.getTransferable().getTransferData(RowSelection.DATA_FLAVOR);
				result = isRowDragAcceptable(dtde, rows);
				if (result) {
					mModel.setDragRows(rows);
				}
			}
		} catch (Exception exception) {
			assert false : Debug.throwableToString(exception);
		}
		return result;
	}

	/**
	 * @param dtde The drop target drag event.
	 * @param column The column.
	 * @return <code>true</code> if the contents of the drag can be dropped into this outline.
	 */
	protected boolean isColumnDragAcceptable(DropTargetDragEvent dtde, Column column) {
		return mModel.getColumns().contains(column);
	}

	/**
	 * @param dtde The drop target drag event.
	 * @param rows The rows.
	 * @return <code>true</code> if the contents of the drag can be dropped into this outline.
	 */
	protected boolean isRowDragAcceptable(DropTargetDragEvent dtde, Row[] rows) {
		return rows.length > 0 && mModel.getRows().contains(rows[0]);
	}

	public void dragEnter(DropTargetDragEvent dtde) {
		mDragWasAcceptable = isDragAcceptable(dtde);
		if (mDragWasAcceptable) {
			Row[] rows;
			if (mModel.getDragColumn() != null) {
				dtde.acceptDrag(dragEnterColumn(dtde));
				return;
			}
			rows = mModel.getDragRows();
			if (rows != null && rows.length > 0) {
				dtde.acceptDrag(dragEnterRow(dtde));
				return;
			}
		}
		dtde.rejectDrag();
	}

	/**
	 * Called when a column drag is entered.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dragEnterColumn(DropTargetDragEvent dtde) {
		mSavedColumns = new ArrayList<Column>(mModel.getColumns());
		return DnDConstants.ACTION_MOVE;
	}

	/**
	 * Called when a row drag is entered.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dragEnterRow(DropTargetDragEvent dtde) {
		addDragHighlight(this);
		return DnDConstants.ACTION_MOVE;
	}

	/**
	 * Called when a row drag is entered over a proxy to this outline.
	 * 
	 * @param dtde The drag event.
	 * @param proxy The proxy.
	 */
	protected void dragEnterRow(DropTargetDragEvent dtde, OutlineProxy proxy) {
		addDragHighlight(proxy);
	}

	public void dragOver(DropTargetDragEvent dtde) {
		if (mDragWasAcceptable) {
			if (mModel.getDragColumn() != null) {
				dtde.acceptDrag(dragOverColumn(dtde));
				return;
			}
			Row[] rows = mModel.getDragRows();
			if (rows != null && rows.length > 0) {
				dtde.acceptDrag(dragOverRow(dtde));
				return;
			}
		}
		dtde.rejectDrag();
	}

	/**
	 * Called when a column drag is in progress.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dragOverColumn(DropTargetDragEvent dtde) {
		int x = dtde.getLocation().x;
		int over = overColumnIndex(x);
		int cur = mModel.getIndexOfColumn(mModel.getDragColumn());
		if (over != cur && over != -1) {
			int midway = getColumnIndexStart(over) + mModel.getColumnAtIndex(over).getWidth() / 2;
			if (over < cur && x < midway || over > cur && x > midway) {
				List<Column> columns = mModel.getColumns();
				if (cur < over) {
					for (int i = cur; i < over; i++) {
						columns.set(i, mModel.getColumnAtIndex(i + 1));
					}
				} else {
					for (int j = cur; j > over; j--) {
						columns.set(j, mModel.getColumnAtIndex(j - 1));
					}
				}
				columns.set(over, mModel.getDragColumn());
				repaint();
				if (mHeaderPanel != null) {
					mHeaderPanel.repaint();
				}
			}
		}
		return DnDConstants.ACTION_MOVE;
	}

	/**
	 * Called when a row drag is in progress.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dragOverRow(DropTargetDragEvent dtde) {
		Row savedParentRow = mDragParentRow;
		int savedChildInsertIndex = mDragChildInsertIndex;
		Row parentRow = null;
		int childInsertIndex = -1;
		Point pt = dtde.getLocation();
		int y = getInsets().top;
		int last = getLastRowToDisplay();
		Row[] dragRows = mModel.getDragRows();
		boolean isFromSelf = dragRows != null && dragRows.length > 0 && mModel.getRows().contains(dragRows[0]);
		Rectangle bounds;
		int indent;
		Row row;

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			row = mModel.getRowAtIndex(i);
			if (!isRowFiltered(row)) {
				int height = row.getHeight();
				if (pt.y <= y + height / 2) {
					if (!isFromSelf || !mModel.isExtendedRowSelected(i) || i != 0 && !mModel.isExtendedRowSelected(i - 1)) {
						parentRow = row.getParent();
						if (parentRow != null) {
							childInsertIndex = parentRow.getIndexOfChild(row);
						} else {
							childInsertIndex = i;
						}
						break;
					}
				} else if (pt.y <= y + height) {
					if (row.canHaveChildren()) {
						bounds = getRowBounds(row);
						indent = mModel.getIndentWidth() + mModel.getIndentWidth(row, mModel.getColumns().get(0));
						if (pt.x >= bounds.x + indent && (!isFromSelf || !mModel.isExtendedRowSelected(row))) {
							parentRow = row;
							childInsertIndex = 0;
							break;
						}
					}
					if (!isFromSelf || !mModel.isExtendedRowSelected(i) || i < last && !mModel.isExtendedRowSelected(i + 1)) {
						parentRow = row.getParent();
						if (parentRow != null) {
							if (!isFromSelf || !mModel.isExtendedRowSelected(i)) {
								childInsertIndex = parentRow.getIndexOfChild(row) + 1;
								break;
							}
						} else {
							childInsertIndex = i + 1;
							break;
						}
					}
				}
				y += height + (mDrawRowDividers ? 1 : 0);
			}
		}
		if (childInsertIndex == -1) {
			if (last > 0) {
				row = mModel.getRowAtIndex(last);
				if (row.canHaveChildren()) {
					bounds = getRowBounds(row);
					indent = mModel.getIndentWidth() + mModel.getIndentWidth(row, mModel.getColumns().get(0));
					if (pt.x >= bounds.x + indent && (!isFromSelf || !mModel.isExtendedRowSelected(row))) {
						parentRow = row;
						childInsertIndex = 0;
					}
				}
				if (childInsertIndex == -1) {
					parentRow = row.getParent();
					if (parentRow != null && (!isFromSelf || !mModel.isExtendedRowSelected(parentRow))) {
						childInsertIndex = parentRow.getIndexOfChild(row) + 1;
					} else {
						parentRow = null;
						childInsertIndex = last + 1;
					}
				}
			} else {
				parentRow = null;
				childInsertIndex = 0;
			}
		}

		if (mDragParentRow != parentRow || mDragChildInsertIndex != childInsertIndex) {
			Graphics gc = getGraphics();
			mDragParentRow = parentRow;
			mDragChildInsertIndex = childInsertIndex;
			drawDragRowInsertionMarker(gc, mDragParentRow, mDragChildInsertIndex);
			gc.dispose();
		}

		if (mDragParentRow != savedParentRow || mDragChildInsertIndex != savedChildInsertIndex) {
			repaint(getDragRowInsertionMarkerBounds(savedParentRow, savedChildInsertIndex));
		}
		return DnDConstants.ACTION_MOVE;
	}

	public void dropActionChanged(DropTargetDragEvent dtde) {
		if (mDragWasAcceptable) {
			Row[] rows;

			if (mModel.getDragColumn() != null) {
				dtde.acceptDrag(dropActionChangedColumn(dtde));
				return;
			}
			rows = mModel.getDragRows();
			if (rows != null && rows.length > 0) {
				dtde.acceptDrag(dropActionChangedRow(dtde));
				return;
			}
		}
		dtde.rejectDrag();
	}

	/**
	 * Called when a column drop action is changed.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dropActionChangedColumn(DropTargetDragEvent dtde) {
		return DnDConstants.ACTION_MOVE;
	}

	/**
	 * Called when a row drop action is changed.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dropActionChangedRow(DropTargetDragEvent dtde) {
		return DnDConstants.ACTION_MOVE;
	}

	public void dragExit(DropTargetEvent dte) {
		if (mDragWasAcceptable) {
			if (mModel.getDragColumn() != null) {
				dragExitColumn(dte);
			} else {
				Row[] rows = mModel.getDragRows();

				if (rows != null && rows.length > 0) {
					dragExitRow(dte);
				}
			}
		}
	}

	/**
	 * Called when a column drag leaves the outline.
	 * 
	 * @param dte The drop target event.
	 */
	protected void dragExitColumn(DropTargetEvent dte) {
		List<Column> columns = mModel.getColumns();

		if (columns.equals(mSavedColumns)) {
			repaintColumn(mModel.getDragColumn());
		} else {
			columns.clear();
			columns.addAll(mSavedColumns);
			repaint();
			if (mHeaderPanel != null) {
				mHeaderPanel.repaint();
			}
		}
		mSavedColumns = null;
		mModel.setDragColumn(null);
	}

	/**
	 * Called when a row drag leaves the outline.
	 * 
	 * @param dte The drop target event.
	 */
	protected void dragExitRow(DropTargetEvent dte) {
		repaint(getDragRowInsertionMarkerBounds(mDragParentRow, mDragChildInsertIndex));
		removeDragHighlight(this);
		mDragParentRow = null;
		mDragChildInsertIndex = -1;
		mModel.setDragRows(null);
	}

	/**
	 * Called when a row drag leaves a proxy of this outline.
	 * 
	 * @param dte The drop target event.
	 * @param proxy The proxy.
	 */
	protected void dragExitRow(DropTargetEvent dte, OutlineProxy proxy) {
		removeDragHighlight(proxy);
	}

	public void drop(DropTargetDropEvent dtde) {
		dtde.acceptDrop(dtde.getDropAction());
		if (mModel.getDragColumn() != null) {
			dropColumn(dtde);
		} else {
			Row[] rows = mModel.getDragRows();

			if (rows != null && rows.length > 0) {
				dropRow(dtde);
			}
		}
		dtde.dropComplete(true);
	}

	/**
	 * Called when a column drag leaves the outline.
	 * 
	 * @param dtde The drop target drop event.
	 */
	protected void dropColumn(DropTargetDropEvent dtde) {
		repaintColumn(mModel.getDragColumn());
		mSavedColumns = null;
		mModel.setDragColumn(null);
	}

	/**
	 * Called to convert the foreign drag rows to this outline's row type and remove them from the
	 * other outline. The default implementation merely removes them from the other outline and
	 * returns the original drag rows, plus any of their children that should be added due to the
	 * row being open. All rows put in the list will have had their owner set to this outline.
	 * 
	 * @param list A list to hold the converted rows.
	 */
	public void convertDragRowsToSelf(List<Row> list) {
		Row[] rows = mModel.getDragRows();
		rows[0].getOwner().removeRows(rows);
		for (Row element : rows) {
			mModel.collectRowsAndSetOwner(list, element, false);
		}
	}

	/**
	 * Called when a row drop occurs.
	 * 
	 * @param dtde The drop target drop event.
	 */
	protected void dropRow(DropTargetDropEvent dtde) {
		removeDragHighlight(this);
		if (mDragChildInsertIndex != -1) {
			StateEdit edit = new StateEdit(mModel, MSG_ROW_DROP_UNDO_TITLE);
			Row[] dragRows = mModel.getDragRows();
			boolean isFromSelf = dragRows != null && dragRows.length > 0 && mModel.getRows().contains(dragRows[0]);
			int count = mModel.getRowCount();
			ArrayList<Row> rows = new ArrayList<Row>(count);
			ArrayList<Row> selection = new ArrayList<Row>(count);
			ArrayList<Row> needSelected = new ArrayList<Row>(count);
			List<Row> modelRows;
			int i;
			int insertAt;
			Row row;

			// Collect up the selected rows
			if (!isFromSelf) {
				convertDragRowsToSelf(selection);
			} else {
				for (i = 0; i < count; i++) {
					row = mModel.getRowAtIndex(i);
					if (mModel.isExtendedRowSelected(row)) {
						selection.add(row);
					}
				}
			}

			// Re-order the visible rows
			if (mDragParentRow != null && !mDragParentRow.isOpen()) {
				insertAt = -1;
			} else {
				insertAt = getAbsoluteInsertionIndex(mDragParentRow, mDragChildInsertIndex);
			}
			for (i = 0; i < count; i++) {
				row = mModel.getRowAtIndex(i);
				if (i == insertAt) {
					rows.addAll(selection);
				}
				if (!isFromSelf || !mModel.isExtendedRowSelected(row)) {
					rows.add(row);
				}
			}
			if (count == insertAt) {
				rows.addAll(selection);
			}

			// Prune the selected rows that don't need to have their parents updated
			needSelected.addAll(selection);
			count = selection.size() - 1;
			for (i = count; i >= 0; i--) {
				Row parent;

				row = selection.get(i);
				parent = row.getParent();
				if (insertAt == -1) {
					row.setOwner(null);
				}
				if (parent != null && (!isFromSelf && !selection.contains(parent) || isFromSelf && !mModel.isExtendedRowSelected(parent))) {
					row.removeFromParent();
				} else if (parent != null) {
					selection.remove(i);
				}
			}

			// Update the parents of the remaining selected rows
			if (mDragParentRow != null) {
				count = selection.size() - 1;
				for (i = count; i >= 0; i--) {
					mDragParentRow.insertChild(mDragChildInsertIndex, selection.get(i));
				}
			}

			mModel.deselect();
			mModel.clearSort();
			mDragParentRow = null;
			mDragChildInsertIndex = -1;
			modelRows = mModel.getRows();
			modelRows.clear();
			modelRows.addAll(rows);
			mModel.getSelection().setSize(modelRows.size());
			setSize(getPreferredSize());
			mModel.select(needSelected, false);
			edit.end();
			postUndo(edit);
			repaint();
			contentSizeMayHaveChanged();
			rowsWereDropped();
		}
		mModel.setDragRows(null);
	}

	/**
	 * Called when a row drop occurs in a proxy of this outline.
	 * 
	 * @param dtde The drop target drop event.
	 * @param proxy The proxy.
	 */
	protected void dropRow(DropTargetDropEvent dtde, OutlineProxy proxy) {
		removeDragHighlight(proxy);
	}

	/** Called after a row drop. */
	protected void rowsWereDropped() {
		// Does nothing.
	}

	private void addDragHighlight(Outline outline) {
		outline.mDragFocus = true;
		outline.repaintFocus();
	}

	private void removeDragHighlight(Outline outline) {
		outline.mDragFocus = false;
		outline.repaintFocus();
	}

	private int getAbsoluteInsertionIndex(Row parent, int childInsertIndex) {
		int insertAt;
		int count;

		if (parent == null) {
			count = mModel.getRowCount();
			insertAt = childInsertIndex;
			while (insertAt < count && mModel.getRowAtIndex(insertAt).getParent() != null) {
				insertAt++;
			}
		} else {
			int i = parent.getChildCount();

			if (i == 0 || !parent.isOpen()) {
				insertAt = mModel.getIndexOfRow(parent) + 1;
			} else if (childInsertIndex < i) {
				insertAt = mModel.getIndexOfRow(parent.getChild(childInsertIndex));
			} else {
				Row row = parent.getChild(i - 1);

				count = mModel.getRowCount();
				insertAt = mModel.getIndexOfRow(row) + 1;
				while (insertAt < count && mModel.getRowAtIndex(insertAt).isDescendentOf(row)) {
					insertAt++;
				}
			}
		}
		return insertAt;
	}

	/**
	 * Causes all row heights to be recalculated, if necessary.
	 * 
	 * @param columns The columns that had their width altered.
	 */
	public void updateRowHeightsIfNeeded(Collection<Column> columns) {
		if (dynamicRowHeight()) {
			for (Column column : columns) {
				if (column.getRowCell(null).participatesInDynamicRowLayout()) {
					updateRowHeights();
					break;
				}
			}
		}
	}

	/** @param row Causes the row height to be recalculated. */
	public void updateRowHeight(Row row) {
		ArrayList<Row> rows = new ArrayList<Row>(1);
		rows.add(row);
		updateRowHeights(rows);
	}

	/** Causes all row heights to be recalculated. */
	public void updateRowHeights() {
		updateRowHeights(mModel.getRows());
	}

	/**
	 * Causes row heights to be recalculated.
	 * 
	 * @param rows The rows to update.
	 */
	public void updateRowHeights(Collection<? extends Row> rows) {
		List<Column> columns = mModel.getColumns();
		boolean needRevalidate = false;

		for (Row row : rows) {
			int height = row.getHeight();
			int prefHeight = row.getPreferredHeight(columns);
			if (height != prefHeight) {
				row.setHeight(prefHeight);
				needRevalidate = true;
			}
		}
		if (needRevalidate) {
			contentSizeMayHaveChanged();
			revalidateView();
		}
	}

	public Insets getAutoscrollInsets() {
		// RAW: Implement
// TKScrollBarOwner scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);
//
// if (scrollPane != null) {
// TKScrollContentView contentView = scrollPane.getContentView();
// Rectangle bounds = contentView.getScrollingBounds(true);
//
// convertRectangle(bounds, contentView, this);
// return new Insets(bounds.y + AUTO_SCROLL_MARGIN, bounds.x + AUTO_SCROLL_MARGIN, getHeight() -
// (bounds.y + bounds.height - AUTO_SCROLL_MARGIN), getWidth() - (bounds.x + bounds.width -
// AUTO_SCROLL_MARGIN));
// }
		return new Insets(0, 0, 0, 0);
	}

	public void autoscroll(Point pt) {
		// RAW: Implement
// TKScrollBarOwner scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);
//
// if (scrollPane != null) {
// // Find the closest edge
// TKScrollContentView contentView = scrollPane.getContentView();
// Rectangle bounds = contentView.getScrollingBounds(true);
//
// convertRectangle(bounds, contentView, this);
//
// int topDelta = pt.y - bounds.y;
// int leftDelta = pt.x - bounds.x;
// int bottomDelta = bounds.y + bounds.height - pt.y;
// int rightDelta = bounds.x + bounds.width - pt.x;
// boolean vertical = topDelta < leftDelta && topDelta < rightDelta || bottomDelta < leftDelta &&
// bottomDelta < rightDelta;
// boolean upLeft = vertical ? topDelta < bottomDelta : leftDelta < rightDelta;
// int x = getX();
// int y = getY();
//
// scrollPane.scroll(vertical, upLeft, false);
// pt.x += x - getX();
// pt.y += y - getY();
// }
	}

	/**
	 * Called whenever the contents of this outline changed due to a user action such that its
	 * preferred size might be different now.
	 */
	public void contentSizeMayHaveChanged() {
		notifyActionListeners(new ActionEvent(getRealOutline(), ActionEvent.ACTION_PERFORMED, getPotentialContentSizeChangeActionCommand()));
	}

	public void rowsAdded(OutlineModel model, Row[] rows) {
		contentSizeMayHaveChanged();
		revalidateView();
	}

	public void rowsWillBeRemoved(OutlineModel model, Row[] rows) {
		// Nothing to do.
	}

	public void rowsWereRemoved(OutlineModel model, Row[] rows) {
		for (Row element : rows) {
			if (element == mRollRow) {
				mRollRow = null;
				break;
			}
		}
		contentSizeMayHaveChanged();
		revalidateView();
	}

	public void sortCleared(OutlineModel model) {
		repaintHeader();
	}

	public void sorted(OutlineModel model, boolean restoring) {
		if (!restoring && isFocusOwner()) {
			scrollSelectionIntoView();
		}
		repaintView();
	}

	/** Scrolls the selection into view, if possible. */
	public void scrollSelectionIntoView() {
		int first = mModel.getFirstSelectedRowIndex();
		if (first >= getFirstRowToDisplay() && first <= getLastRowToDisplay()) {
			scrollSelectionIntoViewInternal();
		} else if (mProxies != null) {
			for (Outline proxy : mProxies) {
				if (first >= proxy.getFirstRowToDisplay() && first <= proxy.getLastRowToDisplay()) {
					proxy.scrollSelectionIntoViewInternal();
					break;
				}
			}
		}
	}

	private void scrollSelectionIntoViewInternal() {
		Selection selection = mModel.getSelection();
		int first = selection.nextSelectedIndex(getFirstRowToDisplay());
		int max = getLastRowToDisplay();

		if (first != -1 && first <= max) {
			Rectangle bounds = getRowIndexBounds(first);
			int tmp = first;
			int last;

			do {
				last = tmp;
				tmp = selection.nextSelectedIndex(last + 1);
			} while (tmp != -1 && tmp <= max);

			if (first != last) {
				bounds = Geometry.union(bounds, getRowIndexBounds(last));
			}
			scrollRectToVisible(bounds);
		}
	}

	public void lockedStateWillChange(OutlineModel model) {
		// Nothing to do...
	}

	public void lockedStateDidChange(OutlineModel model) {
		// Nothing to do...
	}

	public void selectionWillChange(OutlineModel model) {
		repaintSelectionInternal();
	}

	public void selectionDidChange(OutlineModel model) {
		Rectangle bounds = repaintSelectionInternal();

		if (!bounds.isEmpty() && isFocusOwner()) {
			scrollRectToVisible(bounds);
		}
		if (!(this instanceof OutlineProxy)) {
			notifyOfSelectionChange();
		}
	}

	public void undoWillHappen(OutlineModel model) {
		// Nothing to do.
	}

	public void undoDidHappen(OutlineModel model) {
		contentSizeMayHaveChanged();
		revalidateView();
	}

	/** @param undo The undo to post. */
	public void postUndo(UndoableEdit undo) {
		Container top = getTopLevelAncestor();
		if (top instanceof AppWindow) {
			UndoManager mgr = ((AppWindow) top).getUndoManager();
			if (mgr != null) {
				mgr.addEdit(undo);
			}
		}
	}

	/**
	 * @param index The row index to look for.
	 * @return The {@link Outline} most suitable for displaying the index.
	 */
	public Outline getBestOutlineForRowIndex(int index) {
		Outline outline = getRealOutline();

		for (Outline other : outline.mProxies) {
			if (other.mFirstRow <= index && other.mLastRow >= index) {
				return other;
			}
		}
		return outline;
	}

	/**
	 * @param selected Whether or not the selected version of the color is needed.
	 * @param active Whether or not the active version of the color is needed.
	 * @return The foreground color.
	 */
	public static Color getListForeground(boolean selected, boolean active) {
		if (selected) {
			Color color = UIManager.getColor("List.selectionForeground"); //$NON-NLS-1$
			if (!active) {
				Color background = getListBackground(selected, active);
				boolean threshold = Colors.threshold(color, 50);
				if (threshold == Colors.threshold(background, 50)) {
					return threshold ? Color.BLACK : Color.WHITE;
				}
			}
			return color;
		}
		return UIManager.getColor("List.foreground"); //$NON-NLS-1$
	}

	/**
	 * @param selected Whether or not the selected version of the color is needed.
	 * @param active Whether or not the active version of the color is needed.
	 * @return The background color.
	 */
	public static Color getListBackground(boolean selected, boolean active) {
		if (selected) {
			Color color = UIManager.getColor("List.selectionBackground"); //$NON-NLS-1$
			if (!active) {
				color = Colors.lighter(color, 50);
			}
			return color;
		}
		return Color.WHITE;
	}

	public void componentHidden(ComponentEvent event) {
		// Not used.
	}

	public void componentMoved(ComponentEvent event) {
		if (isFocusOwner()) {
			repaint();
		}
	}

	public void componentResized(ComponentEvent event) {
		// Not used.
	}

	public void componentShown(ComponentEvent event) {
		// Not used.
	}
}
