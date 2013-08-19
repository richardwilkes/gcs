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

import com.trollworks.gcs.utility.io.DragUtil;
import com.trollworks.gcs.widgets.GraphicsUtilities;

import java.awt.AlphaComposite;
import java.awt.Color;
import java.awt.Composite;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics;
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
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.util.ArrayList;
import java.util.List;

import javax.swing.JPanel;

/** A header panel for use with {@link Outline}. */
public class OutlineHeader extends JPanel implements DragGestureListener, DropTargetListener, DragSourceListener, MouseListener, MouseMotionListener {
	private Outline	mOwner;
	private Column	mSortColumn;
	private boolean	mResizeOK;
	private boolean	mIgnoreResizeOK;

	/**
	 * Creates a new outline header.
	 * 
	 * @param owner The owning outline.
	 */
	public OutlineHeader(Outline owner) {
		super();
		mOwner = owner;
		setOpaque(true);
		addMouseListener(this);
		addMouseMotionListener(this);

		if (!GraphicsUtilities.inHeadlessPrintMode()) {
			DragSource.getDefaultDragSource().createDefaultDragGestureRecognizer(this, DnDConstants.ACTION_COPY_OR_MOVE, this);
			setDropTarget(new DropTarget(this, this));
		}
// TKOutlineHeaderCM.install();
	}

	public void mouseClicked(MouseEvent event) {
		// Not used
	}

	public void mouseEntered(MouseEvent event) {
		// Not used
	}

	public void mouseExited(MouseEvent event) {
		// Not used
	}

	public void mousePressed(MouseEvent event) {
		if (event.isPopupTrigger()) {
			Column column = mOwner.overColumn(event.getX());

			if (column != null && mOwner.allowColumnContextMenu()) {
				ArrayList<Column> selection = new ArrayList<Column>();

				selection.add(column);
// TKContextMenuManager.showContextMenu(event, this, selection);
			}
		} else if (mOwner.overColumnDivider(event.getX()) == null) {
			mSortColumn = mOwner.overColumn(event.getX());
		} else {
			mOwner.mousePressed(event);
		}
	}

	public void mouseReleased(MouseEvent event) {
		if (mSortColumn != null) {
			if (mSortColumn == mOwner.overColumn(event.getX())) {
				boolean sortAscending = mSortColumn.isSortAscending();

				if (mSortColumn.getSortSequence() != -1) {
					sortAscending = !sortAscending;
				}

				mOwner.setSort(mSortColumn, sortAscending, event.isShiftDown());
			}
			mSortColumn = null;
		} else {
			mOwner.mouseReleased(event);
		}
	}

	public void mouseDragged(MouseEvent event) {
		// Not used
	}

	public void mouseMoved(MouseEvent event) {
		Cursor cursor = Cursor.getDefaultCursor();
		int x = event.getX();
		if (mOwner.overColumnDivider(x) != null) {
			if (mOwner.allowColumnResize()) {
				cursor = Cursor.getPredefinedCursor(Cursor.W_RESIZE_CURSOR);
			}
		} else if (mOwner.overColumn(x) != null) {
			cursor = Cursor.getPredefinedCursor(Cursor.HAND_CURSOR);
		}
		setCursor(cursor);
	}

	/** @return The owning outline. */
	public Outline getOwner() {
		return mOwner;
	}

	@Override public Dimension getPreferredSize() {
		List<Column> columns = mOwner.getModel().getColumns();
		boolean drawDividers = mOwner.shouldDrawColumnDividers();
		Insets insets = getInsets();
		Dimension size = new Dimension(insets.left + insets.right, 0);
		ArrayList<Column> changed = new ArrayList<Column>();

		for (Column col : columns) {
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
		size.height += insets.top + insets.bottom + 1;
		return size;
	}

	@Override protected void paintComponent(Graphics gc) {
		super.paintComponent(GraphicsUtilities.prepare(gc));

		Rectangle clip = gc.getClipBounds();
		Insets insets = getInsets();
		int height = getHeight() - 1;
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), height - (insets.top + insets.bottom));
		boolean drawDividers = mOwner.shouldDrawColumnDividers();
		Color dividerColor = mOwner.getDividerColor();

		gc.setColor(dividerColor);
		gc.drawLine(clip.x, height, clip.x + clip.width, height);
		for (Column col : mOwner.getModel().getColumns()) {
			if (col.isVisible()) {
				bounds.width = col.getWidth();
				if (clip.intersects(bounds)) {
					boolean dragging = mOwner.getSourceDragColumn() == col;
					Composite savedComposite = null;

					if (dragging) {
						savedComposite = ((Graphics2D) gc).getComposite();
						((Graphics2D) gc).setComposite(AlphaComposite.getInstance(AlphaComposite.SRC_OVER, 0.5f));
					}
					col.drawHeaderCell(mOwner, gc, bounds);
					if (dragging) {
						((Graphics2D) gc).setComposite(savedComposite);
					}
				}

				bounds.x += bounds.width;
				if (drawDividers) {
					gc.setColor(dividerColor);
					gc.drawLine(bounds.x, bounds.y, bounds.x, bounds.y + bounds.height);
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

	/**
	 * @param column The column.
	 * @return The bounds of the specified header column.
	 */
	public Rectangle getColumnBounds(Column column) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		bounds.x = mOwner.getColumnStart(column);
		bounds.width = column.getWidth();
		return bounds;
	}

	@Override public String getToolTipText(MouseEvent event) {
		Column column = mOwner.overColumn(event.getX());

		if (column != null) {
			return column.getHeaderCell().getToolTipText(event, getColumnBounds(column), null, column);
		}
		return super.getToolTipText(event);
	}

	public void dragGestureRecognized(DragGestureEvent dge) {
		DragUtil.prepDrag();
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
