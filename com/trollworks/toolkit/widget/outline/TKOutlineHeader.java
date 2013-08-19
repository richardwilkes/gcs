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

import com.trollworks.toolkit.utility.TKDragUtil;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKRectUtils;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.menu.TKContextMenuManager;

import java.awt.AWTEvent;
import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Composite;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DragGestureEvent;
import java.awt.dnd.DragGestureListener;
import java.awt.dnd.DragSource;
import java.awt.dnd.DragSourceDragEvent;
import java.awt.dnd.DragSourceDropEvent;
import java.awt.dnd.DragSourceEvent;
import java.awt.dnd.DragSourceListener;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import java.util.List;

/** A header panel for use with {@link TKOutline}. */
public class TKOutlineHeader extends TKPanel implements DragGestureListener, DropTargetListener, DragSourceListener {
	private TKOutline	mOwner;
	private TKColumn	mSortColumn;
	private boolean		mResizeOK;
	private boolean		mIgnoreResizeOK;

	/**
	 * Creates a new outline header.
	 * 
	 * @param owner The owning outline.
	 */
	public TKOutlineHeader(TKOutline owner) {
		super();
		mOwner = owner;
		setOpaque(true);
		enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);

		if (!TKGraphics.inHeadlessPrintMode()) {
			DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_COPY_OR_MOVE, this);
			setDropTarget(new DropTarget(this, this));
		}
		TKOutlineHeaderCM.install();
	}

	/** @return The owning outline. */
	public TKOutline getOwner() {
		return mOwner;
	}

	@Override protected Dimension getPreferredSizeSelf() {
		List<TKColumn> columns = mOwner.getModel().getColumns();
		boolean drawDividers = mOwner.shouldDrawColumnDividers();
		Insets insets = getInsets();
		Dimension size = new Dimension(insets.left + insets.right, 0);
		ArrayList<TKColumn> changed = new ArrayList<TKColumn>();

		for (TKColumn col : columns) {
			if (col.isVisible()) {
				int tmp = col.getWidth();

				if (tmp == -1) {
					tmp = col.getPreferredWidth(mOwner);
					col.setWidth(tmp);
					changed.add(col);
				}
				size.width += tmp + (drawDividers ? 1 : 0);

				tmp = col.getPreferredHeaderHeight();
				if (tmp > size.height) {
					size.height = tmp;
				}
			}
		}

		if (!changed.isEmpty()) {
			mOwner.updateRowHeightsIfNeeded(changed);
			mOwner.revalidateView();
		}
		size.height += insets.top + insets.bottom;
		return size;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Rectangle bounds = getLocalInsetBounds();
		boolean drawDividers = mOwner.shouldDrawColumnDividers();
		Color dividerColor = mOwner.getDividerColor();

		g2d.setColor(dividerColor);
		for (TKColumn col : mOwner.getModel().getColumns()) {
			if (col.isVisible()) {
				bounds.width = col.getWidth();
				if (TKRectUtils.intersects(clips, bounds)) {
					boolean dragging = mOwner.getSourceDragColumn() == col;
					Composite savedComposite = null;

					if (dragging) {
						savedComposite = g2d.getComposite();
						g2d.setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
					}
					col.drawHeaderCell(g2d, bounds);
					if (dragging) {
						g2d.setComposite(savedComposite);
					}
				}

				bounds.x += bounds.width;
				if (drawDividers) {
					g2d.setColor(dividerColor);
					g2d.drawLine(bounds.x, bounds.y, bounds.x, bounds.y + bounds.height);
					bounds.x++;
				}
			}
		}
	}

	@Override public void repaint(Rectangle bounds) {
		if (mOwner != null) {
			mOwner.repaintHeader(bounds);
		}
	}

	/**
	 * The real version of {@link #repaint(Rectangle)}.
	 * 
	 * @param bounds The bounds to repaint.
	 */
	void repaintInternal(Rectangle bounds) {
		super.repaint(bounds);
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		int x = event.getX();

		switch (event.getID()) {
			case MouseEvent.MOUSE_PRESSED:
				if (TKPopupMenu.isPopupTrigger(event)) {
					TKColumn column = mOwner.overColumn(event.getX());

					if (column != null && mOwner.allowColumnContextMenu()) {
						ArrayList<TKColumn> selection = new ArrayList<TKColumn>();

						selection.add(column);
						TKContextMenuManager.showContextMenu(event, this, selection);
					}
				} else if (mOwner.overColumnDivider(x) == null) {
					mSortColumn = mOwner.overColumn(x);
					mOwner.stopEditing();
				} else {
					mOwner.processMouseEventSelf(event);
				}
				break;
			case MouseEvent.MOUSE_RELEASED:
				if (mSortColumn != null) {
					if (mSortColumn == mOwner.overColumn(x)) {
						boolean sortAscending = mSortColumn.isSortAscending();

						if (mSortColumn.getSortSequence() != -1) {
							sortAscending = !sortAscending;
						}

						mOwner.setSort(mSortColumn, sortAscending, event.isShiftDown());
					}
					mSortColumn = null;
				} else {
					mOwner.processMouseEventSelf(event);
				}
				break;
			case MouseEvent.MOUSE_MOVED:
				Cursor cursor = Cursor.getDefaultCursor();

				if (mOwner.overColumnDivider(x) != null) {
					if (mOwner.allowColumnResize()) {
						cursor = Cursor.getPredefinedCursor(Cursor.W_RESIZE_CURSOR);
					}
				} else if (mOwner.overColumn(x) != null) {
					cursor = Cursor.getPredefinedCursor(Cursor.HAND_CURSOR);
				}
				setCursor(cursor);
				break;
			default:
				if (mSortColumn == null) {
					mOwner.processMouseEventSelf(event);
				}
				break;
		}
	}

	/**
	 * @param column The column.
	 * @return The bounds of the specified header column.
	 */
	public Rectangle getColumnBounds(TKColumn column) {
		Rectangle bounds = getLocalInsetBounds();

		bounds.x = mOwner.getColumnStart(column);
		bounds.width = column.getWidth();
		return bounds;
	}

	@Override public Rectangle getToolTipProhibitedArea(MouseEvent event) {
		TKColumn column = mOwner.overColumn(event.getX());

		if (column != null) {
			return new Rectangle(mOwner.getColumnStart(column), 0, column.getWidth(), getHeight());
		}
		return super.getToolTipProhibitedArea(event);
	}

	@Override public String getToolTipText(MouseEvent event) {
		TKColumn column = mOwner.overColumn(event.getX());

		if (column != null) {
			return column.getHeaderCell().getToolTipText(event, getColumnBounds(column), null, column);
		}
		return super.getToolTipText(event);
	}

	public void dragGestureRecognized(DragGestureEvent dge) {
		TKDragUtil.prepDrag();
		if (mSortColumn != null && mOwner.allowColumnDrag()) {
			mOwner.setSourceDragColumn(mSortColumn);
			if (DragSource.isDragImageSupported()) {
				Point pt = dge.getDragOrigin();

				dge.startDrag(null, mOwner.getColumnDragImage(mSortColumn), new Point(-(pt.x - mOwner.getColumnStart(mSortColumn)), -pt.y), mSortColumn, this);
			} else {
				dge.startDrag(null, mSortColumn, this);
			}
			mSortColumn = null;
		}
	}

	public void dragEnter(DropTargetDragEvent dtde) {
		if (mOwner.getSourceDragColumn() != null) {
			mOwner.dragEnter(dtde);
		} else {
			dtde.rejectDrag();
		}
	}

	public void dragOver(DropTargetDragEvent dtde) {
		if (mOwner.getSourceDragColumn() != null) {
			mOwner.dragOver(dtde);
		} else {
			dtde.rejectDrag();
		}
	}

	public void dropActionChanged(DropTargetDragEvent dtde) {
		if (mOwner.getSourceDragColumn() != null) {
			mOwner.dropActionChanged(dtde);
		} else {
			dtde.rejectDrag();
		}
	}

	public void dragExit(DropTargetEvent dte) {
		if (mOwner.getSourceDragColumn() != null) {
			mOwner.dragExit(dte);
		}
	}

	public void drop(DropTargetDropEvent dtde) {
		mOwner.drop(dtde);
	}

	@Override public void setBounds(int x, int y, int width, int height) {
		if (mIgnoreResizeOK || mResizeOK) {
			super.setBounds(x, y, width, height);
		}
	}

	/** @param resizeOK Whether resizing is allowed or not. */
	void setResizeOK(boolean resizeOK) {
		mResizeOK = resizeOK;
	}

	/** @param ignoreResizeOK Whether {@link #setResizeOK(boolean)} is ignored. */
	public void setIgnoreResizeOK(boolean ignoreResizeOK) {
		mIgnoreResizeOK = ignoreResizeOK;
	}

	public void dragEnter(DragSourceDragEvent dsde) {
		// Nothing to do...
	}

	public void dragOver(DragSourceDragEvent dsde) {
		// Nothing to do...
	}

	public void dropActionChanged(DragSourceDragEvent dsde) {
		// Nothing to do...
	}

	public void dragDropEnd(DragSourceDropEvent dsde) {
		mOwner.setSourceDragColumn(null);
	}

	public void dragExit(DragSourceEvent dse) {
		// Nothing to do...
	}
}
