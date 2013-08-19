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

package com.trollworks.toolkit.widget.outline;

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.undo.TKUndo;
import com.trollworks.toolkit.undo.TKUndoManager;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKDragUtil;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.TKRectUtils;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKSelection;
import com.trollworks.toolkit.widget.menu.TKContextMenuManager;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.widget.scroll.TKScrollBarOwner;
import com.trollworks.toolkit.widget.scroll.TKScrollContentView;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.scroll.TKScrollable;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKKeyDispatcher;
import com.trollworks.toolkit.window.TKUserInputManager;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.AWTEvent;
import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Composite;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
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
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;
import java.util.StringTokenizer;

/** A panel that can show both hierarchical and tabular data. */
public class TKOutline extends TKPanel implements TKOutlineModelListener, ActionListener, FocusListener, Autoscroll, KeyListener, TKScrollable, TKMenuTarget, DragGestureListener, DropTargetListener {
	/** The default double-click action command. */
	public static final String		CMD_OPEN_SELECTION					= "Outline.OpenSelection";				//$NON-NLS-1$
	/** The default selection changed action command. */
	public static final String		CMD_SELECTION_CHANGED				= "Outline.SelectionChanged";			//$NON-NLS-1$
	/** The default delete selection action command. */
	public static final String		CMD_DELETE_SELECTION				= "Outline.DeleteSelection";			//$NON-NLS-1$
	/** The default potential content size change action command. */
	public static final String		CMD_POTENTIAL_CONTENT_SIZE_CHANGE	= "Outline.ContentSizeMayHaveChanged";	//$NON-NLS-1$
	/** The action command to be sent by embedded editors when their contents change. */
	public static final String		CMD_UPDATE_FROM_EDITOR				= "Outline.UpdateFromEditor";			//$NON-NLS-1$
	/** The column visibility command. */
	public static final String		CMD_TOGGLE_COLUMN_VISIBILITY		= "Outline.ToggleColumnVisibility";	//$NON-NLS-1$
	private static final int		DIVIDER_HIT_SLOP					= 2;
	private static final int		AUTO_SCROLL_MARGIN					= 10;
	private TKOutlineModel			mModel;
	/** The header panel. */
	protected TKOutlineHeader		mHeaderPanel;
	private TKRow					mEditRow;
	private TKColumn				mEditColumn;
	private boolean					mForwardToEditor;
	private TKCell					mActiveCell;
	private TKPanel					mEditor;
	private boolean					mNoEditOnKeyboardFocus;
	private boolean					mDrawRowDividers;
	private boolean					mDrawColumnDividers;
	private Color					mDividerColor;
	private boolean					mDrawingDragImage;
	private Rectangle				mDragClip;
	private TKColumn				mDividerDrag;
	private int						mColumnStart;
	private String					mSelectionChangedCommand;
	private String					mDeleteSelectionCommand;
	private String					mPotentialContentSizeChangeCommand;
	private boolean					mAllowColumnContextMenu;
	private boolean					mAllowColumnResize;
	private boolean					mAllowColumnDrag;
	private boolean					mAllowRowDrag;
	private String					mDefaultConfig;
	private boolean					mUseBanding;
	private ArrayList<TKColumn>		mSavedColumns;
	private BufferedImage			mDownTriangle;
	private BufferedImage			mDownTriangleRoll;
	private BufferedImage			mRightTriangle;
	private BufferedImage			mRightTriangleRoll;
	private TKRow					mRollRow;
	private TKRow					mDragParentRow;
	private int						mDragChildInsertIndex;
	private boolean					mDragWasAcceptable;
	private TKColumn				mSourceDragColumn;
	private boolean					mDynamicRowHeight;
	private HashSet<TKProxyOutline>	mProxies;
	/** The first row index this outline will display. */
	protected int					mFirstRow;
	/** The last row index this outline will display. */
	protected int					mLastRow;
	private int						mStartingEdit;
	private int						mSelectOnMouseUp;
	private TKMenuTarget			mMenuTargetDelegate;

	/** Creates a new outline. */
	public TKOutline() {
		this(true);
	}

	/**
	 * Creates a new outline.
	 * 
	 * @param model The model to use.
	 */
	public TKOutline(TKOutlineModel model) {
		this(model, true);
	}

	/**
	 * Creates a new outline.
	 * 
	 * @param showIndent Pass in <code>true</code> if the outline should show hierarchy and
	 *            controls for it.
	 */
	public TKOutline(boolean showIndent) {
		this(new TKOutlineModel(), showIndent);
	}

	/**
	 * Creates a new outline.
	 * 
	 * @param model The model to use.
	 * @param showIndent Pass in <code>true</code> if the outline should show hierarchy and
	 *            controls for it.
	 */
	public TKOutline(TKOutlineModel model, boolean showIndent) {
		super();
		mModel = model;
		mProxies = new HashSet<TKProxyOutline>();
		mAllowColumnContextMenu = true;
		mAllowColumnResize = true;
		mAllowColumnDrag = true;
		mAllowRowDrag = true;
		mDrawRowDividers = true;
		mDrawColumnDividers = true;
		mUseBanding = true;
		mDividerColor = TKColor.SCROLL_BAR_LINE;
		mSelectionChangedCommand = CMD_SELECTION_CHANGED;
		mDeleteSelectionCommand = CMD_DELETE_SELECTION;
		mPotentialContentSizeChangeCommand = CMD_POTENTIAL_CONTENT_SIZE_CHANGE;
		mDownTriangle = TKImage.getDownTriangleIcon();
		mDownTriangleRoll = TKImage.getDownTriangleRollIcon();
		mRightTriangle = TKImage.getRightTriangleIcon();
		mRightTriangleRoll = TKImage.getRightTriangleRollIcon();
		mDragChildInsertIndex = -1;
		mLastRow = -1;
		mModel.setShowIndent(showIndent);
		mModel.setIndentWidth(mDownTriangle.getWidth());

		setActionCommand(CMD_OPEN_SELECTION);
		setBackground(Color.white);
		setOpaque(true);
		setFocusable(true);
		enableAWTEvents(AWTEvent.KEY_EVENT_MASK | AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
		addFocusListener(this);

		if (!TKGraphics.inHeadlessPrintMode()) {
			DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_COPY_OR_MOVE, this);
			setDropTarget(new DropTarget(this, this));
		}

		if (!(this instanceof TKProxyOutline)) {
			mModel.addListener(this);
		}
	}

	/** @return The underlying data model. */
	public TKOutlineModel getModel() {
		return mModel;
	}

	/** @return This outline. */
	public TKOutline getRealOutline() {
		return this;
	}

	/** @return The {@link TKMenuTarget} delegate, if any. */
	public TKMenuTarget getMenuTargetDelegate() {
		return mMenuTargetDelegate;
	}

	/** @param target The new {@link TKMenuTarget} delegate. */
	public void setMenuTargetDelegate(TKMenuTarget target) {
		mMenuTargetDelegate = target;
	}

	/** @param proxy The proxy to add. */
	protected void addProxy(TKProxyOutline proxy) {
		mProxies.add(proxy);
		mModel.addListener(proxy);
	}

	/** Removes all proxies from this outline. */
	public void clearProxies() {
		for (TKProxyOutline proxy : mProxies) {
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

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		Dimension size = new Dimension(insets.left + insets.right, insets.top + insets.bottom);
		List<TKColumn> columns = mModel.getColumns();
		boolean needRevalidate = false;

		for (TKColumn col : columns) {
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

		for (int i = getFirstRowToDisplay(); i <= getLastRowToDisplay(); i++) {
			TKRow row = mModel.getRowAtIndex(i);
			int height = row.getHeight();

			if (height == -1) {
				height = row.getPreferredHeight(columns);
				row.setHeight(height);
			}
			size.height += height + (mDrawRowDividers ? 1 : 0);
		}
		if (mDrawRowDividers && !mModel.getRows().isEmpty()) {
			size.height--;
		}

		return size;
	}

	private void drawDragRowInsertionMarker(Graphics2D g2d, TKRow parent, int insertAtIndex) {
		Rectangle bounds = getDragRowInsertionMarkerBounds(parent, insertAtIndex);

		g2d.setColor(Color.red);
		g2d.drawLine(bounds.x, bounds.y + bounds.height / 2, bounds.x + bounds.width, bounds.y + bounds.height / 2);
		for (int i = 0; i < bounds.height / 2; i++) {
			g2d.drawLine(bounds.x + i, bounds.y + i, bounds.x + i, bounds.y + bounds.height - (1 + i));
		}
	}

	private Rectangle getDragRowInsertionMarkerBounds(TKRow parent, int insertAtIndex) {
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

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Shape origClip = g2d.getClip();
		Rectangle bounds = getLocalInsetBounds();
		Rectangle allClips = TKRectUtils.union(clips);
		boolean active = isFocusOwner();
		int last = getLastRowToDisplay();
		boolean isPrinting = isPrinting();
		boolean showIndent = showIndent();

		for (int rowIndex = getFirstRowToDisplay(); rowIndex <= last; rowIndex++) {
			TKRow row = mModel.getRowAtIndex(rowIndex);

			bounds.height = row.getHeight();
			if (bounds.y >= allClips.y || bounds.y + bounds.height + (mDrawRowDividers ? 1 : 0) >= allClips.y) {
				boolean rowSelected;

				if (bounds.y > allClips.y + allClips.height) {
					break;
				}

				rowSelected = !isPrinting && mModel.isRowSelected(row);
				if (!mDrawingDragImage || mDrawingDragImage && rowSelected) {
					Rectangle[] newClips = TKRectUtils.intersection(clips, bounds);

					if (newClips.length > 0) {
						Rectangle colBounds = new Rectangle(bounds);
						Color rowBackground = getBackground(rowIndex, rowSelected, active);
						Composite savedComposite = null;
						boolean isFirstCol = true;
						int shift = 0;

						for (TKColumn col : mModel.getColumns()) {
							if (col.isVisible()) {
								colBounds.width = col.getWidth();
								if (TKRectUtils.intersects(newClips, colBounds)) {
									boolean dragging = mSourceDragColumn == col;

									if (!mDrawColumnDividers) {
										colBounds.width++;
									}
									g2d.clip(colBounds);
									if (dragging) {
										savedComposite = g2d.getComposite();
										g2d.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
									}
									g2d.setColor(rowBackground);
									g2d.fill(colBounds);
									if (!mDrawColumnDividers) {
										colBounds.width--;
										g2d.clip(colBounds);
									}
									if (showIndent && isFirstCol) {
										shift = mModel.getIndentWidth(row, col);
										colBounds.x += shift;
										colBounds.width -= shift;
										if (row.canHaveChildren()) {
											BufferedImage image = getDisclosureControl(row);

											g2d.drawImage(image, colBounds.x - image.getWidth(), 1 + colBounds.y + (colBounds.height - image.getHeight()) / 2, null);
										}
									}
									// Under some circumstances, the width calculations
									// for cells are off by one pixel when printing...
									// so far, the only way I've found to compensate is
									// to put this hack in.
									if (isPrinting) {
										colBounds.width++;
									}
									col.drawRowCell(g2d, colBounds, row, rowSelected, active);
									if (isPrinting) {
										colBounds.width--;
									}
									if (showIndent && isFirstCol) {
										colBounds.x -= shift;
										colBounds.width += shift;
									}
									if (dragging) {
										g2d.setComposite(savedComposite);
									}
									g2d.setClip(origClip);
								}
								colBounds.x += colBounds.width + 1;
								isFirstCol = false;
							}
						}

						g2d.setColor(rowBackground);
						g2d.fill(new Rectangle(colBounds.x, colBounds.y, bounds.x + bounds.width - colBounds.x, colBounds.height));
					}
					if (mDrawRowDividers) {
						g2d.setColor(mDividerColor);
						g2d.drawLine(bounds.x, bounds.y + bounds.height, bounds.x + bounds.width, bounds.y + bounds.height);
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

		drawColumnDividers(g2d);

		if (mDragChildInsertIndex != -1) {
			drawDragRowInsertionMarker(g2d, mDragParentRow, mDragChildInsertIndex);
		}
	}

	private void drawColumnDividers(Graphics2D g2d) {
		if (mDrawColumnDividers) {
			Rectangle bounds = getLocalInsetBounds();
			int x = bounds.x;
			int top = bounds.y;
			int bottom = top + bounds.height;

			g2d.setColor(mDividerColor);
			for (TKColumn col : mModel.getColumns()) {
				if (col.isVisible()) {
					x += col.getWidth();
					g2d.drawLine(x, top, x, bottom);
					x++;
				}
			}
		}
	}

	@Override public void repaint(Rectangle bounds) {
		super.repaint(bounds);

		// We have to check for null here, since repaint() will be called during
		// initialization of our super class.
		if (mProxies != null) {
			for (TKProxyOutline proxy : mProxies) {
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
			for (TKProxyOutline proxy : mProxies) {
				if (proxy.mHeaderPanel != null) {
					proxy.getHeaderPanel().repaintInternal(bounds);
				}
			}
		}
	}

	/** Repaints the header panel, if present. */
	void repaintHeader() {
		if (mHeaderPanel != null) {
			mHeaderPanel.repaintInternal(mHeaderPanel.getLocalBounds());
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
	public void repaintColumn(TKColumn column) {
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
		for (TKProxyOutline proxy : mProxies) {
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
		Rectangle bounds = getLocalInsetBounds();
		int last = getLastRowToDisplay();
		List<TKColumn> columns = mModel.getColumns();

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			TKRow row = mModel.getRowAtIndex(i);
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
				area = TKRectUtils.union(area, bounds);
			}
			bounds.y += height;
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
			return active ? TKColor.HIGHLIGHT : TKColor.INACTIVE_HIGHLIGHT;
		}
		return useBanding() ? rowIndex % 2 == 0 ? TKColor.PRIMARY_BANDING : TKColor.SECONDARY_BANDING : Color.white;
	}

	public int getBlockScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		return (upLeftDirection ? -1 : 1) * (vertical ? visibleBounds.height : visibleBounds.width);
	}

	public Dimension getPreferredViewportSize() {
		return getPreferredSize();
	}

	public int getUnitScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		if (vertical) {
			Insets insets = getInsets();
			int y = visibleBounds.y - insets.top;
			int rowIndex;
			int rowTop;

			if (upLeftDirection) {
				rowIndex = overRowIndex(y);
				if (rowIndex > -1) {
					rowTop = getRowIndexStart(rowIndex);
					if (rowTop < y) {
						return rowTop - y;
					} else if (--rowIndex > -1) {
						return getRowIndexStart(rowIndex) - y;
					}
				}
			} else {
				y += visibleBounds.height;
				rowIndex = overRowIndex(y);
				if (rowIndex > -1) {
					int rowBottom;

					rowTop = getRowIndexStart(rowIndex);
					rowBottom = rowTop + mModel.getRowAtIndex(rowIndex).getHeight() + (mDrawRowDividers ? 1 : 0);
					if (rowBottom > y) {
						return rowBottom - (y - 1);
					} else if (++rowIndex < mModel.getRowCount()) {
						return getRowIndexStart(rowIndex) + mModel.getRowAtIndex(rowIndex).getHeight() + (mDrawRowDividers ? 1 : 0) - (y - 1);
					}
				}
			}
		}
		return upLeftDirection ? -10 : 10;
	}

	public boolean shouldTrackViewportHeight() {
		return TKScrollPanel.shouldTrackViewportHeight(this);
	}

	public boolean shouldTrackViewportWidth() {
		return TKScrollPanel.shouldTrackViewportWidth(this);
	}

	/**
	 * Determines if the specified x-coordinate is over a column's divider.
	 * 
	 * @param x The coordinate to check.
	 * @return The column, or <code>null</code> if none is found.
	 */
	public TKColumn overColumnDivider(int x) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();

		for (int i = 0; i < count; i++) {
			TKColumn col = mModel.getColumnAtIndex(i);

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
			TKColumn col = mModel.getColumnAtIndex(i);

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
	public TKColumn overColumn(int x) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();

		for (int i = 0; i < count; i++) {
			TKColumn col = mModel.getColumnAtIndex(i);

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
			TKColumn col = mModel.getColumnAtIndex(i);

			if (col.isVisible()) {
				pos += col.getWidth() + (mDrawColumnDividers ? 1 : 0);
				if (x < pos) {
					return i;
				}
			}
		}
		return -1;
	}

	private BufferedImage getDisclosureControl(TKRow row) {
		return row.isOpen() ? row == mRollRow ? mDownTriangleRoll : mDownTriangle : row == mRollRow ? mRightTriangleRoll : mRightTriangle;
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @param column The column the coordinates are currently over.
	 * @param row The row the coordinates are currently over.
	 * @return <code>true</code> if the coordinates are over a disclosure triangle.
	 */
	public boolean overDisclosureControl(int x, @SuppressWarnings("unused") int y, TKColumn column, TKRow row) {
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
			TKColumn column = mModel.getColumnAtIndex(i);

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
	public int getColumnStart(TKColumn column) {
		int pos = getInsets().left;
		int count = mModel.getColumnCount();

		for (int i = 0; i < count; i++) {
			TKColumn col = mModel.getColumnAtIndex(i);

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
	public BufferedImage getColumnDragImage(TKColumn column) {
		BufferedImage offscreen = null;

		synchronized (getTreeLock()) {
			Graphics2D g2d = null;

			try {
				Rectangle bounds = new Rectangle(0, 0, column.getWidth() + (mDrawColumnDividers ? 2 : 0), getHeight() + (mHeaderPanel != null ? mHeaderPanel.getHeight() + 1 : 0));

				offscreen = getGraphicsConfiguration().createCompatibleImage(bounds.width, bounds.height, Transparency.TRANSLUCENT);
				g2d = (Graphics2D) offscreen.getGraphics();
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
				assert false : TKDebug.throwableToString(exception);
			} finally {
				if (g2d != null) {
					g2d.dispose();
				}
			}
		}
		return offscreen;
	}

	/**
	 * Draws a single column.
	 * 
	 * @param g2d The graphics object to use.
	 * @param column The column to draw.
	 * @param bounds The bounds to draw within.
	 */
	public void drawOneColumn(Graphics2D g2d, TKColumn column, Rectangle bounds) {
		Shape oldClip = g2d.getClip();
		Color divColor = getDividerColor();
		int last = getLastRowToDisplay();

		if (mHeaderPanel != null) {
			bounds.height = mHeaderPanel.getHeight();
			g2d.setColor(mHeaderPanel.getBackground());
			g2d.fill(bounds);
			column.drawHeaderCell(g2d, bounds);
			bounds.y += mHeaderPanel.getHeight();
			g2d.setColor(divColor);
			g2d.drawLine(bounds.x, bounds.y, bounds.x + bounds.width, bounds.y);
			bounds.y++;
		}

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			TKRow row = mModel.getRowAtIndex(i);

			bounds.height = row.getHeight();
			g2d.setClip(bounds);
			g2d.setColor(getBackground(i, false, true));
			g2d.fill(bounds);
			column.drawRowCell(g2d, bounds, row, false, true);
			g2d.setClip(oldClip);
			if (mDrawRowDividers) {
				g2d.setColor(divColor);
				g2d.drawLine(bounds.x, bounds.y + bounds.height, bounds.x + bounds.width, bounds.y + bounds.height);
			}
			bounds.y += bounds.height + (mDrawRowDividers ? 1 : 0);
		}
	}

	/**
	 * Determines if the specified y-coordinate is over a row.
	 * 
	 * @param y The coordinate to check.
	 * @return The row, or <code>null</code> if none is found.
	 */
	public TKRow overRow(int y) {
		List<TKRow> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			TKRow row = rows.get(i);

			pos += row.getHeight() + (mDrawRowDividers ? 1 : 0);
			if (y < pos) {
				return row;
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
		List<TKRow> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			TKRow row = rows.get(i);

			pos += row.getHeight() + (mDrawRowDividers ? 1 : 0);
			if (y < pos) {
				return i;
			}
		}
		return -1;
	}

	/**
	 * @param y The coordinate to check.
	 * @return The row index to insert at, from <code>0</code> to
	 *         {@link TKOutlineModel#getRowCount()}.
	 */
	public int getRowInsertionIndex(int y) {
		List<TKRow> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			TKRow row = rows.get(i);
			int height = row.getHeight();
			int tmp = pos + height / 2;

			if (y <= tmp) {
				return i;
			}
			pos += height + (mDrawRowDividers ? 1 : 0);
		}
		return last;
	}

	/**
	 * @param index The index of the row.
	 * @return The starting y-coordinate for the specified row index.
	 */
	public int getRowIndexStart(int index) {
		List<TKRow> rows = mModel.getRows();
		int pos = getInsets().top;

		for (int i = getFirstRowToDisplay(); i < index; i++) {
			pos += rows.get(i).getHeight() + (mDrawRowDividers ? 1 : 0);
		}
		return pos;
	}

	/**
	 * @param row The row.
	 * @return The starting y-coordinate for the specified row.
	 */
	public int getRowStart(TKRow row) {
		List<TKRow> rows = mModel.getRows();
		int pos = getInsets().top;
		int last = getLastRowToDisplay();

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			TKRow oneRow = rows.get(i);

			if (row == oneRow) {
				break;
			}
			pos += oneRow.getHeight() + (mDrawRowDividers ? 1 : 0);
		}
		return pos;
	}

	/**
	 * @param rowIndex The index of the row.
	 * @return The bounds of the row at the specified index.
	 */
	public Rectangle getRowIndexBounds(int rowIndex) {
		Rectangle bounds = getLocalInsetBounds();

		bounds.y = getRowIndexStart(rowIndex);
		bounds.height = mModel.getRowAtIndex(rowIndex).getHeight();
		return bounds;
	}

	/**
	 * @param row The row.
	 * @return The bounds of the specified row.
	 */
	public Rectangle getRowBounds(TKRow row) {
		Rectangle bounds = getLocalInsetBounds();

		bounds.y = getRowStart(row);
		bounds.height = row.getHeight();
		return bounds;
	}

	/**
	 * @param row The row to use.
	 * @param column The column to use.
	 * @return The bounds of the specified cell.
	 */
	public Rectangle getCellBounds(TKRow row, TKColumn column) {
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
	public Rectangle getAdjustedCellBounds(TKRow row, TKColumn column) {
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

	/**
	 * Syncs any cell editor that may be present with the current position of the cell it is
	 * intended to edit.
	 */
	public void syncEditorLocation() {
		if (mEditor != null) {
			syncEditorLocationInternal();
		} else {
			for (TKOutline proxy : new ArrayList<TKProxyOutline>(mProxies)) {
				if (proxy.mEditor != null) {
					proxy.syncEditorLocationInternal();
					return;
				}
			}
		}
	}

	/** @return The current edit row. */
	public TKRow getEditRow() {
		return mEditRow;
	}

	/** @return The current edit column. */
	public TKColumn getEditColumn() {
		return mEditColumn;
	}

	private void syncEditorLocationInternal() {
		Rectangle bounds = getAdjustedCellBounds(mEditRow, mEditColumn);

		repaint(mEditor.getBounds());
		mEditor.setBounds(bounds);
		repaint(bounds);
	}

	/** Sets the width of all visible columns to their preferred width. */
	public void sizeColumnsToFit() {
		ArrayList<TKColumn> columns = new ArrayList<TKColumn>();

		for (TKColumn column : mModel.getColumns()) {
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
	protected void repaintProxyRow(TKRow row) {
		repaint(getRowBounds(row));
		for (TKProxyOutline proxy : mProxies) {
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

	/** Notifies all action listeners of a delete selection request. */
	protected void notifyOfDeleteSelectionRequest() {
		notifyActionListeners(new ActionEvent(getRealOutline(), ActionEvent.ACTION_PERFORMED, getDeleteSelectionActionCommand()));
	}

	/** @return The delete selection action command. */
	public String getDeleteSelectionActionCommand() {
		return mDeleteSelectionCommand;
	}

	/** @param command The delete selection action command. */
	public void setDeleteSelectionActionCommand(String command) {
		mDeleteSelectionCommand = command;
	}

	/** @return The otential user-initiated content size change action command. */
	public String getPotentialContentSizeChangeActionCommand() {
		return mPotentialContentSizeChangeCommand;
	}

	/** @param command The otential user-initiated content size change action command. */
	public void setPotentialContentSizeChangeActionCommand(String command) {
		mPotentialContentSizeChangeCommand = command;
	}

	public void menusWillBeAdjusted() {
		if (mMenuTargetDelegate != null) {
			mMenuTargetDelegate.menusWillBeAdjusted();
		}
	}

	public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (TKWindow.CMD_SELECT_ALL.equals(command)) {
			item.setEnabled(mModel.canSelectAll());
		} else if (CMD_TOGGLE_COLUMN_VISIBILITY.equals(command)) {
			TKColumn column = (TKColumn) item.getUserObject();

			item.setTitle(column.toString().replace('\n', ' '));
			item.setMarked(column.isVisible());
			item.setEnabled(!(mModel.getVisibleColumnCount() == 1 && column.isVisible()));
		} else if (TKOutlineHeaderCM.CMD_SHOW_ALL_COLUMNS.equals(command)) {
			item.setEnabled(mModel.getVisibleColumnCount() != mModel.getColumnCount());
		} else if (TKOutlineHeaderCM.CMD_RESET_COLUMNS.equals(command)) {
			String config = getConfig();
			String original = getDefaultConfig();

			item.setEnabled(!config.equals(original));
		} else if (mMenuTargetDelegate != null) {
			return mMenuTargetDelegate.adjustMenuItem(command, item);
		} else {
			return false;
		}
		return true;
	}

	public void menusWereAdjusted() {
		if (mMenuTargetDelegate != null) {
			mMenuTargetDelegate.menusWereAdjusted();
		}
	}

	public boolean obeyCommand(String command, TKMenuItem item) {
		if (TKWindow.CMD_SELECT_ALL.equals(command)) {
			mModel.select();
		} else if (CMD_TOGGLE_COLUMN_VISIBILITY.equals(command)) {
			TKColumn column = (TKColumn) item.getUserObject();

			column.setVisible(!column.isVisible());
			contentSizeMayHaveChanged();
			revalidateView();
		} else if (TKOutlineHeaderCM.CMD_SHOW_ALL_COLUMNS.equals(command)) {
			for (TKColumn column : mModel.getColumns()) {
				column.setVisible(true);
			}
			contentSizeMayHaveChanged();
			revalidateView();
		} else if (TKOutlineHeaderCM.CMD_RESET_COLUMNS.equals(command)) {
			applyConfig(getDefaultConfig());
			contentSizeMayHaveChanged();
		} else if (mMenuTargetDelegate != null) {
			return mMenuTargetDelegate.obeyCommand(command, item);
		} else {
			return false;
		}
		return true;
	}

	/**
	 * Arranges the columns in the same order as the columns passed in.
	 * 
	 * @param columns The column order.
	 */
	public void setColumnOrder(List<TKColumn> columns) {
		ArrayList<TKColumn> list = new ArrayList<TKColumn>(columns);
		List<TKColumn> cols = mModel.getColumns();

		cols.removeAll(columns);
		list.addAll(cols);
		cols.clear();
		cols.addAll(list);
		repaint();
		getHeaderPanel().repaint();
	}

	/** @return The header panel for this table. */
	public TKOutlineHeader getHeaderPanel() {
		if (mHeaderPanel == null) {
			mHeaderPanel = new TKOutlineHeader(this);
		}
		return mHeaderPanel;
	}

	/** @return The source column being dragged. */
	public TKColumn getSourceDragColumn() {
		return mSourceDragColumn;
	}

	/** @param column The source column being dragged. */
	protected void setSourceDragColumn(TKColumn column) {
		if (mSourceDragColumn != null) {
			repaintColumn(mSourceDragColumn);
		}
		mSourceDragColumn = column;
		if (mSourceDragColumn != null) {
			repaintColumn(mSourceDragColumn);
		}
	}

	@Override protected void processKeyEvent(KeyEvent event) {
		super.processKeyEvent(event);
		if (!event.isConsumed() && isEnabled() && (event.getModifiers() & TKKeystroke.getCommandMask()) == 0) {
			int id = event.getID();

			if (id == KeyEvent.KEY_PRESSED) {
				boolean consume = true;
				boolean shiftDown = event.isShiftDown();
				int max = mModel.getRowCount() - 1;
				int scrollTo = -1;
				int row;
				TKScrollBarOwner scrollPane;

				switch (event.getKeyCode()) {
					case KeyEvent.VK_LEFT:
						for (row = max; row >= 0; row--) {
							if (shiftDown && mModel.isExtendedRowSelected(row) || !shiftDown && mModel.isRowSelected(row)) {
								mModel.getRows().get(row).setOpen(false);
							}
						}
						break;
					case KeyEvent.VK_RIGHT:
						for (row = 0; row < mModel.getRowCount(); row++) {
							if (shiftDown && mModel.isExtendedRowSelected(row) || !shiftDown && mModel.isRowSelected(row)) {
								mModel.getRows().get(row).setOpen(true);
							}
						}
						break;
					case KeyEvent.VK_UP:
						scrollTo = mModel.getSelection().selectUp(shiftDown);
						break;
					case KeyEvent.VK_DOWN:
						scrollTo = mModel.getSelection().selectDown(shiftDown);
						break;
					case KeyEvent.VK_HOME:
						scrollTo = mModel.getSelection().selectToHome(shiftDown);
						break;
					case KeyEvent.VK_END:
						scrollTo = mModel.getSelection().selectToEnd(shiftDown);
						break;
					case KeyEvent.VK_PAGE_UP:
						scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);
						if (scrollPane != null) {
							scrollPane.scroll(!shiftDown, true, true);
						}
						break;
					case KeyEvent.VK_PAGE_DOWN:
						scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);
						if (scrollPane != null) {
							scrollPane.scroll(!shiftDown, false, true);
						}
						break;
					default:
						consume = false;
						break;
				}

				if (consume) {
					if (scrollTo != -1) {
						keyScroll(scrollTo);
					}
					event.consume();
				}
			} else if (id == KeyEvent.KEY_TYPED) {
				char ch = event.getKeyChar();

				if (ch == '\n' || ch == '\r') {
					if (mModel.hasSelection()) {
						notifyActionListeners();
					}
					event.consume();
				} else if (ch == '\b' || ch == KeyEvent.VK_DELETE) {
					if (mModel.hasSelection()) {
						notifyOfDeleteSelectionRequest();
					}
					event.consume();
				}
			}
		}
	}

	/** @param scrollTo The row index to scroll to. */
	protected void keyScroll(int scrollTo) {
		TKOutline real = getRealOutline();

		if (!keyScrollInternal(real, scrollTo)) {
			for (TKProxyOutline proxy : real.mProxies) {
				if (keyScrollInternal(proxy, scrollTo)) {
					break;
				}
			}
		}
	}

	private boolean keyScrollInternal(TKOutline outline, int scrollTo) {
		if (scrollTo >= outline.mFirstRow && scrollTo <= outline.mLastRow) {
			outline.requestFocus();
			outline.scrollRectIntoView(outline.getRowIndexBounds(scrollTo));
			return true;
		}
		return false;
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		TKRow rollRow = null;

		try {
			boolean local = event.getSource() == this;
			int id = event.getID();
			int x = event.getX();
			int y = event.getY();
			TKRow rowHit;
			int rowIndexHit;
			TKColumn column;

			switch (id) {
				case MouseEvent.MOUSE_CLICKED:
					int clickCount = event.getClickCount();

					rollRow = mRollRow;
					if (clickCount == 1) {
						if (local) {
							column = overColumn(x);
							if (column != null) {
								rowHit = overRow(y);
								if (rowHit != null && !overDisclosureControl(x, y, column, rowHit)) {
									TKCell cell = column.getRowCell(rowHit);

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
										boolean okToNotify = true;

										if (column != null) {
											TKCell cell = column.getRowCell(rowHit);

											if (cell != null && cell.isEditable(rowHit, column, false)) {
												startEditing(rowHit, column, null);
												mForwardToEditor = true;
												okToNotify = false;
											}
										}
										if (okToNotify) {
											notifyActionListeners();
										}
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
					break;
				case MouseEvent.MOUSE_PRESSED:
					setNoEditOnNextKeyboardFocus();
					requestFocus();
					stopEditing();
					mSelectOnMouseUp = -1;
					mActiveCell = null;
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
							mActiveCell = column.getRowCell(rowHit);
						}
						if (mActiveCell != null && mActiveCell.isEditable(rowHit, column, true)) {
							startEditing(rowHit, column, event);
							mForwardToEditor = true;
						} else {
							int method = TKSelection.MOUSE_NONE;

							rowIndexHit = overRowIndex(y);

							if (event.isShiftDown()) {
								method |= TKSelection.MOUSE_EXTEND;
							}
							if (TKKeystroke.isCommandKeyDown(event) && !TKPopupMenu.isPopupTrigger(event)) {
								method |= TKSelection.MOUSE_FLIP;
							}
							mSelectOnMouseUp = mModel.getSelection().selectByMouse(rowIndexHit, method);
							if (TKPopupMenu.isPopupTrigger(event)) {
								mSelectOnMouseUp = -1;
								showContextMenu(event);
							}
						}
					}
					break;
				case MouseEvent.MOUSE_RELEASED:
					rollRow = mRollRow;
					if (mForwardToEditor) {
						mForwardToEditor = false;
						if (mEditor != null) {
							TKUserInputManager.forwardMouseEvent(event, this, mEditor);
						}
					} else {
						if (mDividerDrag != null && allowColumnResize()) {
							dragColumnDivider(x);
						}
						mDividerDrag = null;
						if (mSelectOnMouseUp != -1) {
							mModel.select(mSelectOnMouseUp, false);
							mSelectOnMouseUp = -1;
						}
					}
					break;
				case MouseEvent.MOUSE_MOVED:
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
									TKCell cell = column.getRowCell(rowHit);

									cursor = cell.getCursor(event, getCellBounds(rowHit, column), rowHit, column);
								}
							}
						}
					}
					setCursor(cursor);
					break;
				case MouseEvent.MOUSE_DRAGGED:
					mSelectOnMouseUp = -1;
					if (mForwardToEditor) {
						if (mEditor != null) {
							TKUserInputManager.forwardMouseEvent(event, this, mEditor);
						}
					} else if (!TKContextMenuManager.isHandlingContextMenu() && mDividerDrag != null && allowColumnResize()) {
						TKScrollBarOwner scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);

						dragColumnDivider(x);

						if (scrollPane != null) {
							Point pt = event.getPoint();

							if (!(event.getSource() instanceof TKOutline)) {
								// Column resizing is occurring in the header, most likely
								pt.y = getVisibleBounds().y + 1;
							}
							scrollPane.scrollPointIntoView(event, this, pt);
						}
					}
					break;
			}
		} finally {
			if (rollRow != mRollRow) {
				TKColumn column = mModel.getColumnAtIndex(0);
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
	}

	/**
	 * @param viewPt The location within the view.
	 * @return The row cell at the specified point.
	 */
	public TKCell getCellAt(Point viewPt) {
		return getCellAt(viewPt.x, viewPt.y);
	}

	/**
	 * @param x The x-coordinate within the view.
	 * @param y The y-coordinate within the view.
	 * @return The row cell at the specified coordinates.
	 */
	public TKCell getCellAt(int x, int y) {
		TKColumn column = overColumn(x);

		if (column != null) {
			TKRow row = overRow(y);

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
	public void adjustColumnWidth(TKColumn column, int width) {
		ArrayList<TKColumn> columns = new ArrayList<TKColumn>(1);

		column.setWidth(width);
		columns.add(column);
		processColumnWidthChanges(columns);
	}

	private void processColumnWidthChanges(ArrayList<TKColumn> columns) {
		updateRowHeightsIfNeeded(columns);
		revalidateView();
		syncEditorLocation();
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The drag image for this table when dragging rows.
	 */
	@SuppressWarnings("null") protected BufferedImage getDragImage(int x, int y) {
		Graphics2D g2d = null;
		BufferedImage off1 = null;
		BufferedImage off2 = null;

		mDrawingDragImage = true;
		mDragClip = null;
		off1 = getImage(null);
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
			g2d.drawImage(off1, -mDragClip.x, -mDragClip.y, this);
		} catch (Exception paintException) {
			assert false : TKDebug.throwableToString(paintException);
			off2 = null;
			mDragClip = new Rectangle(x, y, 1, 1);
		} finally {
			if (g2d != null) {
				g2d.dispose();
			}
		}

		return off2 != null ? off2 : off1;
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
		TKContextMenuManager.showContextMenu(event, getRealOutline(), mModel.getSelectionAsList());
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
		}
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
	public void setSort(TKColumn column, boolean ascending, boolean add) {
		TKOutlineModelUndoSnapshot before = new TKOutlineModelUndoSnapshot(mModel);
		int count = mModel.getColumnCount();
		int i;

		if (!add) {
			for (i = 0; i < count; i++) {
				TKColumn col = mModel.getColumnAtIndex(i);

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
		postUndo(new TKOutlineModelUndo(Msgs.SORT_UNDO_TITLE, mModel, before, new TKOutlineModelUndoSnapshot(mModel)));
	}

	public void keyPressed(KeyEvent event) {
		if (event.getKeyCode() == KeyEvent.VK_TAB) {
			transferInternalFocus(!event.isShiftDown());
			event.consume();
		}
	}

	public void keyReleased(KeyEvent event) {
		// Nothing to do
	}

	public void keyTyped(KeyEvent event) {
		// Nothing to do
	}

	/**
	 * Changes the focus to the next/previous editable cell, or moves it to the next/previous
	 * focusable component.
	 * 
	 * @param forward Whether or not the focus is moving forward.
	 */
	public void transferInternalFocus(boolean forward) {
		TKRow row = mEditRow;
		TKColumn column = mEditColumn;
		int dir = forward ? 1 : -1;
		int start = mModel.getIndexOfColumn(column);
		int i = start + dir;
		int max = mModel.getColumnCount() - 1;
		int rmax = getLastRowToDisplay();
		int rmin = getFirstRowToDisplay();
		int ri = -2;
		TKCell cell;

		while (i != start) {
			if (i < 0) {
				i = max;
				if (ri == -2) {
					ri = mModel.getIndexOfRow(row);
				}
				if (--ri < rmin) {
					break;
				}
				row = mModel.getRowAtIndex(ri);
			} else if (i > max) {
				i = 0;
				if (ri == -2) {
					ri = mModel.getIndexOfRow(row);
				}
				if (++ri > rmax) {
					break;
				}
				row = mModel.getRowAtIndex(ri);
			}

			column = mModel.getColumnAtIndex(i);
			cell = column.getRowCell(row);
			if (cell.isEditable(row, column, false)) {
				startEditing(row, column, null);
				return;
			}
			i += dir;
		}

		if (forward) {
			KeyboardFocusManager.getCurrentKeyboardFocusManager().focusNextComponent(mEditor != null ? mEditor : this);
		} else {
			KeyboardFocusManager.getCurrentKeyboardFocusManager().focusPreviousComponent(this);
		}
	}

	/**
	 * Causes the next keyboard focus event to not cause editing or initial row selection.
	 */
	public void setNoEditOnNextKeyboardFocus() {
		mNoEditOnKeyboardFocus = true;
	}

	public void focusGained(FocusEvent event) {
		boolean allowEditing = !mNoEditOnKeyboardFocus;
		int rowCount = mModel.getRowCount();

		mNoEditOnKeyboardFocus = false;
		repaintSelection();

		if (allowEditing && !mModel.hasSelection() && rowCount > 0) {
			int first = getFirstRowToDisplay();

			if (rowCount > first) {
				boolean forward = !TKKeyDispatcher.isKeyPressed(KeyEvent.VK_SHIFT);

				startEditing(mModel.getRowAtIndex(forward ? first : getLastRowToDisplay()), forward);
			}
		}
	}

	public void focusLost(FocusEvent event) {
		repaintSelection();
		mNoEditOnKeyboardFocus = false;
		if (!event.isTemporary() && event.getOppositeComponent() != mEditor) {
			stopEditing();
		}
	}

	/**
	 * @param outline The outline to check.
	 * @return Whether the specified outline refers to this outline or a proxy of it.
	 */
	public boolean isSelfOrProxy(TKOutline outline) {
		TKOutline self = getRealOutline();

		if (outline == self) {
			return true;
		}
		for (TKProxyOutline proxy : self.mProxies) {
			if (outline == proxy) {
				return true;
			}
		}
		return false;
	}

	@Override public void windowFocus(boolean gained) {
		super.windowFocus(gained);
		repaintSelection();
	}

	@Override public Rectangle getToolTipProhibitedArea(MouseEvent event) {
		TKColumn column = overColumn(event.getX());

		if (column != null) {
			TKRow row = overRow(event.getY());

			if (row != null) {
				return getCellBounds(row, column);
			}
		}
		return super.getToolTipProhibitedArea(event);
	}

	@Override public String getToolTipText(MouseEvent event) {
		TKColumn column = overColumn(event.getX());

		if (column != null) {
			TKRow row = overRow(event.getY());

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
	public static String createConfig(TKColumnConfig[] configSpec) {
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
	public static String createConfig(TKColumnConfig[] configSpec, int hSplit, int vSplit) {
		StringBuilder buffer = new StringBuilder();

		buffer.append(TKOutlineModel.CONFIG_VERSION);
		buffer.append('\t');
		buffer.append(configSpec.length);
		for (TKColumnConfig element : configSpec) {
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

		buffer.append(TKOutlineModel.CONFIG_VERSION);
		buffer.append('\t');
		buffer.append(count);
		for (int i = 0; i < count; i++) {
			TKColumn column = mModel.getColumnAtIndex(i);

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
			if (getInteger(tokenizer, 0) == TKOutlineModel.CONFIG_VERSION) {
				int count = getInteger(tokenizer, 0);
				List<TKColumn> columns = mModel.getColumns();
				boolean needSort = false;
				TKColumn column;
				int i;

				mModel.clearSort();

				for (i = 0; i < count; i++) {
					column = mModel.getColumnWithID(getInteger(tokenizer, 0));
					if (column == null) {
						throw new Exception();
					}
					columns.remove(column);
					columns.add(i, column);
					column.setVisible(TKNumberUtils.getBoolean(tokenizer.nextToken()));
					column.setWidth(getInteger(tokenizer, column.getWidth()));
					column.setSortCriteria(getInteger(tokenizer, -1), TKNumberUtils.getBoolean(tokenizer.nextToken()));
					if (column.getSortSequence() != -1) {
						needSort = true;
					}
				}
				if (needSort) {
					mModel.sort();
				}
				updateRowHeightsIfNeeded(columns);
				revalidateView();
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
		TKDragUtil.prepDrag();
		if (mDividerDrag == null && mModel.hasSelection() && allowRowDrag() && mActiveCell != null) {
			Point pt = dge.getDragOrigin();
			int index = overRowIndex(pt.y);
			TKRow row = mModel.getRowAtIndex(index);
			TKColumn column = mModel.getColumnAtIndex(overColumnIndex(pt.x));

			if (mActiveCell.isRowDragHandle(row, column)) {
				TKRowSelection selection = new TKRowSelection(mModel, mModel.getSelectionAsList(true).toArray(new TKRow[0]));

				if (DragSource.isDragImageSupported()) {
					dge.startDrag(null, getDragImage(pt.x, pt.y), new Point(mDragClip.x - pt.x, mDragClip.y - pt.y), selection, null);
				} else {
					dge.startDrag(null, selection);
				}
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
			if (dtde.isDataFlavorSupported(TKColumn.DATA_FLAVOR)) {
				TKColumn column = (TKColumn) dtde.getTransferable().getTransferData(TKColumn.DATA_FLAVOR);

				result = isColumnDragAcceptable(dtde, column);
				if (result) {
					mModel.setDragColumn(column);
				}
			}
			if (dtde.isDataFlavorSupported(TKRowSelection.DATA_FLAVOR)) {
				TKRow[] rows = (TKRow[]) dtde.getTransferable().getTransferData(TKRowSelection.DATA_FLAVOR);

				result = isRowDragAcceptable(dtde, rows);
				if (result) {
					mModel.setDragRows(rows);
				}
			}
		} catch (Exception exception) {
			assert false : TKDebug.throwableToString(exception);
		}
		return result;
	}

	/**
	 * @param dtde The drop target drag event.
	 * @param column The column.
	 * @return <code>true</code> if the contents of the drag can be dropped into this outline.
	 */
	protected boolean isColumnDragAcceptable(@SuppressWarnings("unused") DropTargetDragEvent dtde, TKColumn column) {
		return mModel.getColumns().contains(column);
	}

	/**
	 * @param dtde The drop target drag event.
	 * @param rows The rows.
	 * @return <code>true</code> if the contents of the drag can be dropped into this outline.
	 */
	protected boolean isRowDragAcceptable(@SuppressWarnings("unused") DropTargetDragEvent dtde, TKRow[] rows) {
		return rows.length > 0 && mModel.getRows().contains(rows[0]);
	}

	public void dragEnter(DropTargetDragEvent dtde) {
		mDragWasAcceptable = isDragAcceptable(dtde);
		if (mDragWasAcceptable) {
			TKRow[] rows;

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
	protected int dragEnterColumn(@SuppressWarnings("unused") DropTargetDragEvent dtde) {
		mSavedColumns = new ArrayList<TKColumn>(mModel.getColumns());
		return DnDConstants.ACTION_MOVE;
	}

	/**
	 * Called when a row drag is entered.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dragEnterRow(@SuppressWarnings("unused") DropTargetDragEvent dtde) {
		return DnDConstants.ACTION_MOVE;
	}

	/**
	 * Called when a row drag is entered over a proxy to this outline.
	 * 
	 * @param dtde The drag event.
	 * @param proxy The proxy.
	 */
	protected void dragEnterRow(@SuppressWarnings("unused") DropTargetDragEvent dtde, @SuppressWarnings("unused") TKProxyOutline proxy) {
		// Does nothing by default.
	}

	public void dragOver(DropTargetDragEvent dtde) {
		if (mDragWasAcceptable) {
			TKRow[] rows;

			if (mModel.getDragColumn() != null) {
				dtde.acceptDrag(dragOverColumn(dtde));
				return;
			}
			rows = mModel.getDragRows();
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
				List<TKColumn> columns = mModel.getColumns();

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
		TKRow savedParentRow = mDragParentRow;
		int savedChildInsertIndex = mDragChildInsertIndex;
		TKRow parentRow = null;
		int childInsertIndex = -1;
		Point pt = dtde.getLocation();
		int y = getInsets().top;
		int last = getLastRowToDisplay();
		TKRow[] dragRows = mModel.getDragRows();
		boolean isFromSelf = dragRows != null && dragRows.length > 0 && mModel.getRows().contains(dragRows[0]);
		Rectangle bounds;
		int indent;
		TKRow row;

		for (int i = getFirstRowToDisplay(); i <= last; i++) {
			int height;

			row = mModel.getRowAtIndex(i);
			height = row.getHeight();

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
			Graphics2D g2d = getGraphics2D();

			mDragParentRow = parentRow;
			mDragChildInsertIndex = childInsertIndex;
			drawDragRowInsertionMarker(g2d, mDragParentRow, mDragChildInsertIndex);
			g2d.dispose();
		}

		if (mDragParentRow != savedParentRow || mDragChildInsertIndex != savedChildInsertIndex) {
			repaint(getDragRowInsertionMarkerBounds(savedParentRow, savedChildInsertIndex));
		}
		return DnDConstants.ACTION_MOVE;
	}

	public void dropActionChanged(DropTargetDragEvent dtde) {
		if (mDragWasAcceptable) {
			TKRow[] rows;

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
	protected int dropActionChangedColumn(@SuppressWarnings("unused") DropTargetDragEvent dtde) {
		return DnDConstants.ACTION_MOVE;
	}

	/**
	 * Called when a row drop action is changed.
	 * 
	 * @param dtde The drag event.
	 * @return The value to return via {@link DropTargetDragEvent#acceptDrag(int)}.
	 */
	protected int dropActionChangedRow(@SuppressWarnings("unused") DropTargetDragEvent dtde) {
		return DnDConstants.ACTION_MOVE;
	}

	public void dragExit(DropTargetEvent dte) {
		if (mDragWasAcceptable) {
			if (mModel.getDragColumn() != null) {
				dragExitColumn(dte);
			} else {
				TKRow[] rows = mModel.getDragRows();

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
	protected void dragExitColumn(@SuppressWarnings("unused") DropTargetEvent dte) {
		List<TKColumn> columns = mModel.getColumns();

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
	protected void dragExitRow(@SuppressWarnings("unused") DropTargetEvent dte) {
		repaint(getDragRowInsertionMarkerBounds(mDragParentRow, mDragChildInsertIndex));
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
	protected void dragExitRow(@SuppressWarnings("unused") DropTargetEvent dte, @SuppressWarnings("unused") TKProxyOutline proxy) {
		// Does nothing by default.
	}

	public void drop(DropTargetDropEvent dtde) {
		dtde.acceptDrop(dtde.getDropAction());
		if (mModel.getDragColumn() != null) {
			dropColumn(dtde);
		} else {
			TKRow[] rows = mModel.getDragRows();

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
	protected void dropColumn(@SuppressWarnings("unused") DropTargetDropEvent dtde) {
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
	protected void convertDragRowsToSelf(List<TKRow> list) {
		TKRow[] rows = mModel.getDragRows();

		rows[0].getOwner().removeRows(rows);
		for (TKRow element : rows) {
			mModel.collectRowsAndSetOwner(list, element, false);
		}
	}

	/**
	 * Called when a row drop occurs.
	 * 
	 * @param dtde The drop target drop event.
	 */
	protected void dropRow(@SuppressWarnings("unused") DropTargetDropEvent dtde) {
		if (mDragChildInsertIndex != -1) {
			TKOutlineModelUndoSnapshot before = new TKOutlineModelUndoSnapshot(mModel);
			TKRow[] dragRows = mModel.getDragRows();
			boolean isFromSelf = dragRows != null && dragRows.length > 0 && mModel.getRows().contains(dragRows[0]);
			int count = mModel.getRowCount();
			ArrayList<TKRow> rows = new ArrayList<TKRow>(count);
			ArrayList<TKRow> selection = new ArrayList<TKRow>(count);
			ArrayList<TKRow> needSelected = new ArrayList<TKRow>(count);
			List<TKRow> modelRows;
			int i;
			int insertAt;
			TKRow row;

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
				TKRow parent;

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
			setSize(getPreferredSizeSelf());
			mModel.select(needSelected, false);
			postUndo(new TKOutlineModelUndo(Msgs.ROW_DROP_UNDO_TITLE, mModel, before, new TKOutlineModelUndoSnapshot(mModel)));
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
	protected void dropRow(@SuppressWarnings("unused") DropTargetDropEvent dtde, @SuppressWarnings("unused") TKProxyOutline proxy) {
		// Does nothing by default.
	}

	/** Called after a row drop. */
	protected void rowsWereDropped() {
		// Does nothing.
	}

	private int getAbsoluteInsertionIndex(TKRow parent, int childInsertIndex) {
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
				TKRow row = parent.getChild(i - 1);

				count = mModel.getRowCount();
				insertAt = mModel.getIndexOfRow(row) + 1;
				while (insertAt < count && mModel.getRowAtIndex(insertAt).isDescendentOf(row)) {
					insertAt++;
				}
			}
		}
		return insertAt;
	}

	/** @return The current editor, if any. */
	public TKPanel getCurrentEditor() {
		if (mEditor == null) {
			for (TKOutline proxy : new ArrayList<TKProxyOutline>(mProxies)) {
				if (proxy.mEditor != null) {
					return proxy.mEditor;
				}
			}
			return null;
		}
		return mEditor;
	}

	/**
	 * Start editing the first cell in the row that is editable. If no cells are editable, select
	 * the row instead.
	 * 
	 * @param row The row to start editing.
	 * @param forward Pass in <code>true</code> to start looking for the first editable field from
	 *            the first column or <code>false</code> to start from the last column and work
	 *            backwards.
	 */
	public void startEditing(TKRow row, boolean forward) {
		int count = mModel.getColumnCount();
		int start = forward ? 0 : count - 1;
		int end = forward ? count - 1 : 0;
		int inc = forward ? 1 : -1;

		for (int i = start; i != end; i += inc) {
			TKColumn column = mModel.getColumnAtIndex(i);
			TKCell cell = column.getRowCell(row);

			if (cell.isEditable(row, column, false)) {
				startEditing(row, column, null);
				return;
			}
		}
		mModel.select(row, false);
	}

	/**
	 * Start editing the specified row and column. If the cell at that position is not editable,
	 * then nothing more than a call to <code>deselect()</code> occurs.
	 * 
	 * @param row The row to start editing.
	 * @param column The column to start editing.
	 * @param event Pass in the mouse event, if any, that is triggering the start of editing.
	 */
	public void startEditing(TKRow row, TKColumn column, MouseEvent event) {
		if (mEditColumn == column && mEditRow == row && mEditor != null && mEditor.getParent() == this) {
			return;
		}
		mStartingEdit = 1;
		mModel.deselect();
		stopEditing(); // In case deselect didn't call stopEditing()
		if (row != null && column != null && getBaseWindow() != null) {
			TKCell cell = column.getRowCell(row);

			if (cell != null && cell.isEditable(row, column, false)) {
				int index = mModel.getIndexOfRow(row);

				if (index < getFirstRowToDisplay() || index > getLastRowToDisplay()) {
					TKOutline outline = getRealOutline();

					if (index >= outline.getFirstRowToDisplay() && index <= outline.getLastRowToDisplay()) {
						if (event != null) {
							TKUserInputManager.translateMouseEvent(event, this, outline);
						}
						outline.startEditing(row, column, event);
					} else {
						for (TKProxyOutline proxy : outline.mProxies) {
							if (index >= proxy.getFirstRowToDisplay() && index <= proxy.getLastRowToDisplay()) {
								if (event != null) {
									TKUserInputManager.translateMouseEvent(event, this, proxy);
								}
								proxy.startEditing(row, column, event);
								break;
							}
						}
					}
				} else {
					Rectangle bounds = getAdjustedCellBounds(row, column);

					mActiveCell = cell;
					mEditRow = row;
					mEditColumn = column;
					mEditor = cell.getEditor(getBackground(mModel.getIndexOfRow(row), false, true), row, column, event != null);
					mEditor.setFocusTraversalKeysEnabled(false);
					mEditor.addKeyListener(this);
					mEditor.setBounds(bounds);
					add(mEditor);
					mEditor.addActionListener(this);
					mEditor.addFocusListener(new FocusListener() {

						public void focusGained(FocusEvent focusEvent) {
							// Not used
						}

						public void focusLost(FocusEvent focusEvent) {
							if (!focusEvent.isTemporary() && focusEvent.getOppositeComponent() != getCurrentEditor()) {
								stopEditing();
							}
						}
					});
					mEditor.requestFocus();
					mEditor.paintImmediately();
					if (event != null) {
						TKUserInputManager.forwardMouseEvent(event, this, mEditor);
					}
				}
			}
		}
		if (mStartingEdit == 2) {
			contentSizeMayHaveChanged();
		}
		mStartingEdit = 0;
	}

	/** Stop editing. */
	public void stopEditing() {
		stopEditingInternal();
		for (TKProxyOutline proxy : new ArrayList<TKProxyOutline>(mProxies)) {
			proxy.stopEditingInternal();
		}
	}

	/** Stop editing. */
	protected void stopEditingInternal() {
		if (mEditor != null) {
			TKPanel editor = mEditor;
			Rectangle bounds = getCellBounds(mEditRow, mEditColumn);

			mEditor.removeActionListener(this);
			mEditor.removeKeyListener(this);
			mEditor.removeFocusListener(this);
			mEditRow.setData(mEditColumn, mActiveCell.stopEditing(mEditor));
			mEditor = null;
			mActiveCell = null;
			mForwardToEditor = false;
			EventQueue.invokeLater(new DelayedRemover(editor, bounds));
			if (dynamicRowHeight()) {
				Rectangle rowBounds = getRowBounds(mEditRow);
				int prefHeight = mEditRow.getPreferredHeight(mModel.getColumns());

				if (rowBounds.height != prefHeight) {
					mEditRow.setHeight(prefHeight);
					if (mStartingEdit == 1) {
						mStartingEdit = 2;
					} else {
						contentSizeMayHaveChanged();
					}
					revalidate();
				}
			}
		}
	}

	/**
	 * Causes all row heights to be recalculated, if necessary.
	 * 
	 * @param columns The columns that had their width altered.
	 */
	public void updateRowHeightsIfNeeded(Collection<TKColumn> columns) {
		if (dynamicRowHeight()) {
			for (TKColumn column : columns) {
				if (column.getRowCell(null).participatesInDynamicRowLayout()) {
					updateRowHeights();
					break;
				}
			}
		}
	}

	/** @param row Causes the row height to be recalculated. */
	public void updateRowHeight(TKRow row) {
		ArrayList<TKRow> rows = new ArrayList<TKRow>(1);

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
	public void updateRowHeights(Collection<? extends TKRow> rows) {
		List<TKColumn> columns = mModel.getColumns();
		boolean needRevalidate = false;

		for (TKRow row : rows) {
			int height = row.getHeight();
			int prefHeight = row.getPreferredHeight(columns);

			if (height != prefHeight) {
				row.setHeight(prefHeight);
				needRevalidate = true;
			}
		}
		if (needRevalidate) {
			contentSizeMayHaveChanged();
			revalidate();
		}
	}

	/**
	 * @return The index location of the current cell editor, or <code>null</code> if there is
	 *         none.
	 */
	public Point getCellLocationOfEditor() {
		Point result = getCellLocationOfEditorInternal();

		if (result == null) {
			for (TKProxyOutline proxy : mProxies) {
				result = proxy.getCellLocationOfEditorInternal();
				if (result != null) {
					return result;
				}
			}
		}
		return result;
	}

	/**
	 * @return The index location of the current cell editor, or <code>null</code> if there is
	 *         none.
	 */
	protected Point getCellLocationOfEditorInternal() {
		if (mEditor != null) {
			return new Point(mModel.getIndexOfColumn(mEditColumn), mModel.getIndexOfRow(mEditRow));
		}
		return null;
	}

	public Insets getAutoscrollInsets() {
		TKScrollBarOwner scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);

		if (scrollPane != null) {
			TKScrollContentView contentView = scrollPane.getContentView();
			Rectangle bounds = contentView.getScrollingBounds(true);

			convertRectangle(bounds, contentView, this);
			return new Insets(bounds.y + AUTO_SCROLL_MARGIN, bounds.x + AUTO_SCROLL_MARGIN, getHeight() - (bounds.y + bounds.height - AUTO_SCROLL_MARGIN), getWidth() - (bounds.x + bounds.width - AUTO_SCROLL_MARGIN));
		}
		return new Insets(0, 0, 0, 0);
	}

	public void autoscroll(Point pt) {
		TKScrollBarOwner scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);

		if (scrollPane != null) {
			// Find the closest edge
			TKScrollContentView contentView = scrollPane.getContentView();
			Rectangle bounds = contentView.getScrollingBounds(true);

			convertRectangle(bounds, contentView, this);

			int topDelta = pt.y - bounds.y;
			int leftDelta = pt.x - bounds.x;
			int bottomDelta = bounds.y + bounds.height - pt.y;
			int rightDelta = bounds.x + bounds.width - pt.x;
			boolean vertical = topDelta < leftDelta && topDelta < rightDelta || bottomDelta < leftDelta && bottomDelta < rightDelta;
			boolean upLeft = vertical ? topDelta < bottomDelta : leftDelta < rightDelta;
			int x = getX();
			int y = getY();

			scrollPane.scroll(vertical, upLeft, false);
			pt.x += x - getX();
			pt.y += y - getY();
		}
	}

	public void actionPerformed(ActionEvent event) {
		if (CMD_UPDATE_FROM_EDITOR.equals(event.getActionCommand()) && mEditor != null) {
			mEditRow.setData(mEditColumn, mActiveCell.getEditedObject(mEditor));
			mModel.clearSort();
		}
	}

	/**
	 * Called whenever the contents of this outline changed due to a user action such that its
	 * preferred size might be different now.
	 */
	public void contentSizeMayHaveChanged() {
		notifyActionListeners(new ActionEvent(getRealOutline(), ActionEvent.ACTION_PERFORMED, getPotentialContentSizeChangeActionCommand()));
	}

	/** @return A contextual menu for the column. */
	public TKMenu getColumnsMenu() {
		TKMenu menu = new TKMenu(Msgs.COLUMN_MENU_TITLE);

		for (TKColumn column : mModel.getColumns()) {
			TKMenuItem item = new TKMenuItem(column.toString().replace('\n', ' '), CMD_TOGGLE_COLUMN_VISIBILITY);

			item.setUserObject(column);
			menu.add(item);
		}
		menu.addSeparator();
		menu.add(new TKMenuItem(Msgs.SHOW_ALL_COLUMNS_TITLE, TKOutlineHeaderCM.CMD_SHOW_ALL_COLUMNS));
		menu.add(new TKMenuItem(Msgs.RESET_COLUMNS_TITLE, TKOutlineHeaderCM.CMD_RESET_COLUMNS));
		return menu;
	}

	public void rowsAdded(TKOutlineModel model, TKRow[] rows) {
		contentSizeMayHaveChanged();
		revalidate();
	}

	public void rowsWillBeRemoved(TKOutlineModel model, TKRow[] rows) {
		for (TKRow element : rows) {
			if (element == mEditRow) {
				stopEditing();
				break;
			}
		}
	}

	public void rowsWereRemoved(TKOutlineModel model, TKRow[] rows) {
		for (TKRow element : rows) {
			if (element == mRollRow) {
				mRollRow = null;
				break;
			}
		}
		contentSizeMayHaveChanged();
		revalidate();
	}

	public void sortCleared(TKOutlineModel model) {
		repaintHeader();
	}

	public void sorted(TKOutlineModel model, boolean restoring) {
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
			for (TKOutline proxy : mProxies) {
				if (first >= proxy.getFirstRowToDisplay() && first <= proxy.getLastRowToDisplay()) {
					proxy.scrollSelectionIntoViewInternal();
					break;
				}
			}
		}
	}

	private void scrollSelectionIntoViewInternal() {
		TKSelection selection = mModel.getSelection();
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
				bounds = TKRectUtils.union(bounds, getRowIndexBounds(last));
			}
			scrollRectIntoView(bounds);
		}
	}

	public void lockedStateWillChange(TKOutlineModel model) {
		// Nothing to do...
	}

	public void lockedStateDidChange(TKOutlineModel model) {
		// Nothing to do...
	}

	public void selectionWillChange(TKOutlineModel model) {
		stopEditingInternal();
		repaintSelectionInternal();
	}

	public void selectionDidChange(TKOutlineModel model) {
		Rectangle bounds = repaintSelectionInternal();

		if (!bounds.isEmpty() && isFocusOwner()) {
			scrollRectIntoView(bounds);
		}
		if (!(this instanceof TKProxyOutline)) {
			notifyOfSelectionChange();
		}
	}

	public void undoWillHappen(TKOutlineModel model) {
		stopEditing();
	}

	public void undoDidHappen(TKOutlineModel model) {
		contentSizeMayHaveChanged();
		revalidate();
	}

	/** @param undo The undo to post. */
	public void postUndo(TKUndo undo) {
		TKBaseWindow window = getBaseWindow();

		if (window != null) {
			TKUndoManager mgr = window.getUndoManager();

			if (mgr != null) {
				mgr.addEdit(undo);
			}
		}
	}

	private class DelayedRemover implements Runnable {
		private TKPanel		mPanelToRemove;
		private Rectangle	mBounds;

		/**
		 * @param panel The panel to remove.
		 * @param bounds The rectangle to repaint.
		 */
		DelayedRemover(TKPanel panel, Rectangle bounds) {
			mPanelToRemove = panel;
			mBounds = bounds;
		}

		public void run() {
			remove(mPanelToRemove);
			repaint(mBounds);
		}
	}
}
