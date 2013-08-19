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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.collections.TKRangedSize;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.utility.TKTiming;
import com.trollworks.toolkit.widget.menu.TKContextMenuManager;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.widget.scroll.TKScrollBarOwner;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.scroll.TKScrollable;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.AWTEvent;
import java.awt.Color;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.List;

/**
 * Provides a visual list of items.
 * 
 * @param <T> The type of object to present a list of.
 */
public class TKItemList<T> extends TKPanel implements TKMenuTarget, TKScrollable, TKSelection.Owner, FocusListener {
	/** The default double-click action command. */
	public static final String		CMD_OPEN_SELECTION		= "ListOpenSelection";		//$NON-NLS-1$
	/** The default selection changed action command. */
	public static final String		CMD_SELECTION_CHANGED	= "ListSelectionChanged";	//$NON-NLS-1$
	private String					mSelectionChangedCommand;
	private TKItemRenderer			mRenderer;
	private ArrayList<T>			mData;
	private ArrayList<Dimension>	mCachedSizes;
	private TKRangedSize			mHeights;
	private TKSelection				mSelection;
	private boolean					mInMouseDown;
	private boolean					mMultipleSelectionAllowed;
	private boolean					mForceInactive;
	private boolean					mSelectionAllowed;
	private boolean					mHandleContextMenu;
	private int						mSelectOnMouseUp;

	/** Creates an empty list. */
	public TKItemList() {
		super();
		mData = new ArrayList<T>();
		initialize();
	}

	/**
	 * Creates a list with the specified data.
	 * 
	 * @param array The initial set of objects.
	 */
	public TKItemList(T[] array) {
		super();
		mData = new ArrayList<T>(Arrays.asList(array));
		initialize();
	}

	/**
	 * Creates a list with the specified data.
	 * 
	 * @param collection The initial set of objects.
	 */
	public TKItemList(Collection<T> collection) {
		super();
		mData = new ArrayList<T>(collection);
		initialize();
	}

	private void initialize() {
		int count = mData.size();

		mCachedSizes = new ArrayList<Dimension>(count);
		mHandleContextMenu = false;
		mRenderer = new TKDefaultItemRenderer();
		mHeights = new TKRangedSize(0);
		mSelection = new TKSelection(this, count);
		mMultipleSelectionAllowed = true;
		mSelectionAllowed = true;
		mSelectionChangedCommand = CMD_SELECTION_CHANGED;
		setActionCommand(CMD_OPEN_SELECTION);
		setOpaque(true);
		setCursor(Cursor.getDefaultCursor());
		setBackground(Color.white);
		setFocusable(true);
		for (int i = 0; i < count; i++) {
			mCachedSizes.add(new Dimension(-1, -1));
		}
		enableAWTEvents(AWTEvent.KEY_EVENT_MASK | AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
		addFocusListener(this);
	}

	/**
	 * Adds an item to the list.
	 * 
	 * @param item The item to add.
	 */
	public void addItem(T item) {
		if (item != null) {
			mData.add(item);
			mCachedSizes.add(new Dimension(-1, -1));
			adjustForAdd();
		}
	}

	/**
	 * Adds an array of items to the list.
	 * 
	 * @param items The items to add.
	 */
	public void addItems(T[] items) {
		if (items != null) {
			for (T element : items) {
				mData.add(element);
				mCachedSizes.add(new Dimension(-1, -1));
			}
			adjustForAdd();
		}
	}

	/**
	 * Adds a list of items to the list.
	 * 
	 * @param items The items to add.
	 */
	public void addItems(Collection<T> items) {
		if (items != null) {
			int count = items.size();

			mData.addAll(items);
			for (int i = 0; i < count; i++) {
				mCachedSizes.add(new Dimension(-1, -1));
			}
			adjustForAdd();
		}
	}

	private void adjustForAdd() {
		mSelection.setSize(mData.size());
		revalidateImmediately();
		mSelection.deselect();
	}

	/** @return The action command for this list when the selection changes. */
	public String getSelectionChangedActionCommand() {
		return mSelectionChangedCommand;
	}

	/** @param command The action command for this list when the selection changes. */
	public void setSelectionChangedActionCommand(String command) {
		mSelectionChangedCommand = command;
	}

	private void notifyOfSelectionChange() {
		notifyActionListeners(new ActionEvent(this, ActionEvent.ACTION_PERFORMED, getSelectionChangedActionCommand()));
	}

	/**
	 * Adds the specified range to the current selection. If multiple selection is not allowed, this
	 * behaves the same as a call to {@link #setSelection(int)}.
	 * 
	 * @param startIndex The starting index.
	 * @param endIndex The ending index.
	 */
	public void addRangeToSelection(int startIndex, int endIndex) {
		if (isMultipleSelectionAllowed()) {
			mSelection.select(startIndex, endIndex, true);
		} else {
			mSelection.select(startIndex, false);
		}
	}

	/**
	 * Adds a particular index to the current selection. If multiple selection is not allowed, this
	 * behaves the same as a call to {@link #setSelection(int)}.
	 * 
	 * @param index The index to add.
	 */
	public void addToSelection(int index) {
		addRangeToSelection(index, index);
	}

	/**
	 * Adds a particular object to the current selection. If multiple selection is not allowed, this
	 * behaves the same as a call to {@link #setSelection(Object)}.
	 * 
	 * @param obj The object to add to the selection.
	 */
	public void addToSelection(Object obj) {
		int index = mData.indexOf(obj);

		if (index >= 0) {
			addRangeToSelection(index, index);
		}
	}

	public void menusWillBeAdjusted() {
		// Nothing to do...
	}

	public boolean adjustMenuItem(String command, TKMenuItem item) {
		boolean processed = true;

		if (TKWindow.CMD_SELECT_ALL.equals(command)) {
			boolean enabled;

			if (isMultipleSelectionAllowed()) {
				enabled = mSelection.canSelectAll();
			} else {
				enabled = getItemCount() == 1 && !isSelected(0);
			}
			item.setEnabled(enabled);
		} else {
			processed = false;
		}
		return processed;
	}

	public void menusWereAdjusted() {
		// Nothing to do...
	}

	public boolean obeyCommand(String command, TKMenuItem item) {
		boolean processed = true;

		if (TKWindow.CMD_SELECT_ALL.equals(command)) {
			selectAll();
		} else {
			processed = false;
		}
		return processed;
	}

	/** Flashes the selected nodes. */
	public void flashSelectedNodes() {
		if (mSelection.getCount() > 0) {
			TKTiming timing = new TKTiming();

			for (int i = 0; i < 4; i++) {
				if (i != 0) {
					timing.delayUntilThenReset(100);
				}
				mForceInactive = !mForceInactive;
				repaintSelection();
				updateImmediately();
			}
		}
	}

	/**
	 * @param index The index of the item.
	 * @return The item at the specified index, or <code>null</code> if there is none.
	 */
	public T getItem(int index) {
		if (index >= 0 && index < getItemCount()) {
			return mData.get(index);
		}
		return null;
	}

	/** @return The items in this list. */
	public List<T> getItems() {
		return Collections.unmodifiableList(mData);
	}

	/**
	 * @param index The index of the item.
	 * @return The bounding rectangle of the item at the specified index.
	 */
	public Rectangle getItemBounds(int index) {
		if (index >= 0 && index < getItemCount()) {
			Insets insets = getInsets();

			return new Rectangle(insets.left, insets.top + mHeights.getSizeBeforeIndex(index), getWidth() - (insets.left + insets.right), mHeights.getSizeOfIndex(index));
		}
		return new Rectangle();
	}

	/** @return The number of items in this list. */
	public int getItemCount() {
		return mData.size();
	}

	/**
	 * @return The bounds of the focused item in this list. If there is no focused item, then the
	 *         bounds of the entire list is returned.
	 */
	public Rectangle getKeyboardFocusBounds() {
		int which = getSelectedIndex();

		if (which != -1) {
			return getItemBounds(which);
		}
		return getLocalBounds();
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		int count = getItemCount();
		int width = 0;
		int height = 0;

		for (int i = 0; i < count; i++) {
			Dimension cachedSize = mCachedSizes.get(i);

			if (cachedSize.width == -1) {
				Dimension size = mRenderer.getItemPreferredSize(getItem(i), i);

				cachedSize.width = size.width;
				cachedSize.height = size.height;
			}

			if (cachedSize.width > width) {
				width = cachedSize.width;
			}
			height += cachedSize.height;
			mHeights.setSizeOfRange(i, i, cachedSize.height);
		}

		return new Dimension(width + insets.left + insets.right, height + insets.top + insets.bottom);
	}

	/** @return The first selected index, or -1 if there is none. */
	public int getSelectedIndex() {
		return mSelection.firstSelectedIndex();
	}

	/** @return An array of selected indexes. */
	public int[] getSelectedIndexes() {
		return mSelection.getSelectedIndexes();
	}

	/** @return The first selected item. */
	public T getSelectedItem() {
		int index = mSelection.firstSelectedIndex();

		if (index != -1) {
			return getItem(index);
		}
		return null;
	}

	/** @return An array of selected items. */
	public List<T> getSelectedItems() {
		ArrayList<T> selection = new ArrayList<T>();
		int index = mSelection.firstSelectedIndex();

		while (index != -1) {
			selection.add(getItem(index));
			index = mSelection.nextSelectedIndex(index + 1);
		}
		return selection;
	}

	/** @return The number of items selected. */
	public int getSelectionCount() {
		return mSelection.getCount();
	}

	public Dimension getPreferredViewportSize() {
		return getPreferredSize();
	}

	public int getBlockScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		return (upLeftDirection ? -1 : 1) * (vertical ? visibleBounds.height : visibleBounds.width);
	}

	public int getUnitScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		if (vertical) {
			Insets insets = getInsets();
			int y;
			int index;
			Rectangle itemBounds;

			if (upLeftDirection) {
				y = visibleBounds.y - insets.top;
				index = mHeights.getIndexAtPosition(y);
				itemBounds = getItemBounds(index);
				if (itemBounds.y < y) {
					return itemBounds.y - y;
				} else if (index > 0) {
					itemBounds = getItemBounds(index - 1);
					return itemBounds.y - y;
				}
				return -1;
			}

			y = visibleBounds.y + visibleBounds.height - insets.top;
			index = mHeights.getIndexAtPosition(y);
			if (index == -1) {
				return 1;
			}
			itemBounds = getItemBounds(index);
			if (itemBounds.y + itemBounds.height > y) {
				return itemBounds.y + itemBounds.height - y;
			} else if (index < getItemCount() - 1) {
				itemBounds = getItemBounds(index + 1);
				return itemBounds.y + itemBounds.height - y;
			}
			return 1;
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
	 * Inserts an item into the list at the specified index.
	 * 
	 * @param index The index to insert at.
	 * @param item The item to insert.
	 */
	public void insertItem(int index, T item) {
		if (index >= 0 && index <= getItemCount()) {
			mData.add(index, item);
			mCachedSizes.add(index, new Dimension(-1, -1));
			adjustForAdd();
		}
	}

	/**
	 * Inserts an array of items into the list at the specified index.
	 * 
	 * @param index The index to insert at.
	 * @param items The items to insert.
	 */
	public void insertItems(int index, T[] items) {
		if (index >= 0 && index <= getItemCount()) {
			for (T element : items) {
				mData.add(index, element);
				mCachedSizes.add(index++, new Dimension(-1, -1));
			}
			adjustForAdd();
		}
	}

	/**
	 * Inserts a collection of items into the list at the specified index.
	 * 
	 * @param index The index to insert at.
	 * @param items The items to insert.
	 */
	public void insertItems(int index, Collection<T> items) {
		if (index >= 0 && index <= getItemCount()) {
			for (T item : items) {
				mData.add(index, item);
				mCachedSizes.add(index++, new Dimension(-1, -1));
			}
			adjustForAdd();
		}
	}

	/** Invalidates the size cache. Call this if the underlying data has been modified. */
	public void invalidateSizeCache() {
		for (Dimension size : mCachedSizes) {
			size.width = -1;
			size.height = -1;
		}
	}

	/** @return <code>true</code> if more than one item may be selected. */
	public boolean isMultipleSelectionAllowed() {
		return mMultipleSelectionAllowed;
	}

	/** @param allowed Whether multiple items may be selected. */
	public void setMultipleSelectionAllowed(boolean allowed) {
		if (mMultipleSelectionAllowed != allowed) {
			if (!allowed && mSelection.getCount() > 1) {
				mSelection.select(getSelectedIndex(), false);
			}
			mMultipleSelectionAllowed = allowed;
		}
	}

	/**
	 * @param index The index to check.
	 * @return <code>true</code> if the specified index is selected.
	 */
	public boolean isSelected(int index) {
		return mSelectionAllowed && mSelection.isSelected(index);
	}

	public void focusGained(FocusEvent event) {
		if (!mInMouseDown) {
			scrollRectIntoView(getKeyboardFocusBounds());
		}
		repaintKeyboardFocus();
	}

	public void focusLost(FocusEvent event) {
		repaintKeyboardFocus();
	}

	@Override public void windowFocus(boolean gained) {
		super.windowFocus(gained);
		repaintKeyboardFocus();
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Color savedColor = g2d.getColor();
		Rectangle clipBounds = g2d.getClipBounds();
		int bottom = clipBounds.y + clipBounds.height;
		Insets insets = getInsets();
		int count = getItemCount();
		int i = mHeights.getIndexAtPosition(clipBounds.y - insets.top);
		Rectangle bounds = new Rectangle(insets.left, 0, getWidth() - (insets.left + insets.right), 0);
		boolean active = !mForceInactive && isFocusOwner();

		if (i < 0) {
			i = 0;
		}

		for (; i < count; i++) {
			bounds.y = insets.top + mHeights.getSizeBeforeIndex(i);
			if (bounds.y <= bottom) {
				T item = getItem(i);
				boolean selected = isSelected(i);

				bounds.height = mHeights.getSizeOfIndex(i);
				g2d.setColor(mRenderer.getBackgroundForItem(item, i, selected, active));
				g2d.fill(bounds);
				mRenderer.drawItem(g2d, bounds, item, i, selected, active);
			} else {
				break;
			}
		}
		g2d.setColor(savedColor);
	}

	@Override public void processKeyEvent(KeyEvent event) {
		super.processKeyEvent(event);
		if (!event.isConsumed() && mSelectionAllowed && isEnabled()) {
			boolean shiftIsDown = (event.getModifiers() & InputEvent.SHIFT_MASK) != 0;

			switch (event.getID()) {
				case KeyEvent.KEY_PRESSED:
					boolean consume = true;
					int scrollTo = -1;
					TKScrollBarOwner scrollPane;

					switch (event.getKeyCode()) {
						case KeyEvent.VK_UP:
							scrollTo = mSelection.selectUp(mMultipleSelectionAllowed && shiftIsDown);
							break;
						case KeyEvent.VK_DOWN:
							scrollTo = mSelection.selectDown(mMultipleSelectionAllowed && shiftIsDown);
							break;
						case KeyEvent.VK_HOME:
							scrollTo = mSelection.selectToHome(mMultipleSelectionAllowed && shiftIsDown);
							break;
						case KeyEvent.VK_END:
							scrollTo = mSelection.selectToEnd(mMultipleSelectionAllowed && shiftIsDown);
							break;
						case KeyEvent.VK_PAGE_UP:
							scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);
							if (scrollPane != null) {
								scrollPane.scroll(!shiftIsDown, true, true);
							}
							break;
						case KeyEvent.VK_PAGE_DOWN:
							scrollPane = (TKScrollBarOwner) getAncestorOfType(TKScrollBarOwner.class);
							if (scrollPane != null) {
								scrollPane.scroll(!shiftIsDown, false, true);
							}
							break;
						default:
							consume = false;
							break;
					}

					if (consume) {
						if (scrollTo != -1) {
							scrollRectIntoView(getItemBounds(scrollTo));
						}
						event.consume();
					}
					break;
			}
		}
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		if (mSelectionAllowed) {
			switch (event.getID()) {
				case MouseEvent.MOUSE_CLICKED:
					if (event.getClickCount() == 2 && mSelection.getCount() > 0) {
						notifyActionListeners();
					}
					break;
				case MouseEvent.MOUSE_PRESSED:
					Insets insets = getInsets();
					int modifiers = event.getModifiers();
					int hit = mHeights.getIndexAtPosition(event.getY() - insets.top);
					int method = TKSelection.MOUSE_NONE;

					mSelectOnMouseUp = -1;
					mInMouseDown = true;
					requestFocus();

					if (mMultipleSelectionAllowed && (modifiers & InputEvent.SHIFT_MASK) != 0) {
						method |= TKSelection.MOUSE_EXTEND;
					}
					if (TKKeystroke.isCommandKeyDown(event) && !TKPopupMenu.isPopupTrigger(event)) {
						method |= TKSelection.MOUSE_FLIP;
					}
					mSelectOnMouseUp = mSelection.selectByMouse(hit, method);

					if (mHandleContextMenu && TKPopupMenu.isPopupTrigger(event)) {
						mSelectOnMouseUp = -1;
						showContextMenu(event);
					}
					break;
				case MouseEvent.MOUSE_RELEASED:
					mInMouseDown = false;
					if (mSelectOnMouseUp != -1) {
						mSelection.select(mSelectOnMouseUp, false);
						mSelectOnMouseUp = -1;
					}
					break;
				case MouseEvent.MOUSE_DRAGGED:
					mSelectOnMouseUp = -1;
					break;
			}
		}
	}

	/**
	 * Displays a context menu.
	 * 
	 * @param event The mouse event that triggered the contextual menu.
	 */
	protected void showContextMenu(MouseEvent event) {
		TKContextMenuManager.showContextMenu(event, this, getSelectedItems());
	}

	/**
	 * Removes the item at the specified index from the list.
	 * 
	 * @param index The index to remove.
	 */
	public void removeItem(int index) {
		if (index >= 0 && index < getItemCount()) {
			mSelection.deselect();
			mData.remove(index);
			mCachedSizes.remove(index);
			revalidateImmediately();
		}
	}

	/**
	 * Removes a range of items from the list.
	 * 
	 * @param startIndex The starting index to remove.
	 * @param endIndex The ending index to remove.
	 */
	public void removeItems(int startIndex, int endIndex) {
		mSelection.deselect();
		if (startIndex > endIndex) {
			int tmp = startIndex;

			startIndex = endIndex;
			endIndex = tmp;
		}

		if (startIndex < 0) {
			startIndex = 0;
		}
		if (endIndex >= getItemCount()) {
			endIndex = getItemCount() - 1;
		}

		for (int i = endIndex; i >= startIndex; i--) {
			mData.remove(i);
			mCachedSizes.remove(i);
		}
		mSelection.setSize(mData.size());
		revalidateImmediately();
	}

	/**
	 * Removes an item from the list.
	 * 
	 * @param item The item to remove.
	 */
	public void removeItem(T item) {
		int index = mData.indexOf(item);

		if (index != -1) {
			removeItem(index);
		}
	}

	/** Removes all items from the list. */
	public void removeAllItems() {
		mSelection.deselect();
		mSelection.setSize(0);
		mData.clear();
		mCachedSizes.clear();
		revalidateImmediately();
	}

	/**
	 * Removes the specified index from the current selection.
	 * 
	 * @param index The index to remove.
	 */
	public void removeItemFromSelection(int index) {
		if (index >= 0 && index < getItemCount()) {
			mSelection.deselect(index);
		}
	}

	/** Repaints the keyboard focus, if any. */
	protected void repaintKeyboardFocus() {
		if (getSelectedIndex() != -1) {
			repaint(getKeyboardFocusBounds());
		}
	}

	/**
	 * Replaces the existing list items with the specified item.
	 * 
	 * @param item The new item.
	 */
	public void replaceItem(T item) {
		mSelection.deselect();
		mData.clear();
		mCachedSizes.clear();
		if (item != null) {
			mData.add(item);
			mCachedSizes.add(new Dimension(-1, -1));
		}
		mSelection.setSize(mData.size());
		revalidateImmediately();
	}

	/**
	 * Replaces the existing list items with the array of items.
	 * 
	 * @param items The new items.
	 */
	public void replaceItems(T[] items) {
		mSelection.deselect();
		mData.clear();
		mCachedSizes.clear();
		if (items != null) {
			for (T element : items) {
				mData.add(element);
				mCachedSizes.add(new Dimension(-1, -1));
			}
		}
		mSelection.setSize(mData.size());
		revalidateImmediately();
	}

	/**
	 * Replaces the existing list items with the collection of items.
	 * 
	 * @param items The new items.
	 */
	public void replaceItems(Collection<T> items) {
		mSelection.deselect();
		mData.clear();
		mCachedSizes.clear();
		if (items != null) {
			int count = items.size();

			mData.addAll(items);
			for (int i = 0; i < count; i++) {
				mCachedSizes.add(new Dimension(-1, -1));
			}
		}
		mSelection.setSize(mData.size());
		revalidateImmediately();
	}

	/**
	 * Selects all items in the list. If multiple selection is not allowed and there is more than
	 * one item in the list, nothing happens.
	 */
	public void selectAll() {
		if (isMultipleSelectionAllowed() || getItemCount() == 1) {
			mSelection.select();
		}
	}

	/** @param allowed Whether selection is allowed at all. */
	public void setSelectionAllowed(boolean allowed) {
		if (allowed != mSelectionAllowed) {
			mSelectionAllowed = allowed;
			repaint();
		}
	}

	/** @param renderer The item renderer for this list. */
	public void setItemRenderer(TKItemRenderer renderer) {
		mRenderer = renderer;
		repaint();
	}

	/** @param enabled Whether context menu handling should be enabled. */
	public void setHandleContextMenu(boolean enabled) {
		mHandleContextMenu = enabled;
	}

	/** @return <code>true</code> if context menu handling is enabled. */
	public boolean getHandleContextMenu() {
		return mHandleContextMenu;
	}

	/**
	 * Selects a particular item, deselecting any others that may be selected.
	 * 
	 * @param item The item to select.
	 */
	public void setSelection(T item) {
		int which = mData.indexOf(item);

		if (which >= 0) {
			mSelection.select(which, false);
		}
	}

	/**
	 * Selects a particular index, deselecting any others that may be selected.
	 * 
	 * @param index The index to select.
	 */
	public void setSelection(int index) {
		mSelection.select(index, false);
	}

	/** Repaints the current selection. */
	public void repaintSelection() {
		int index = mSelection.firstSelectedIndex();

		while (index != -1) {
			repaint(getItemBounds(index));
			index = mSelection.nextSelectedIndex(index + 1);
		}
	}

	public void selectionAboutToChange() {
		repaintSelection();
	}

	public void selectionDidChange() {
		repaintSelection();
		notifyOfSelectionChange();
	}
}
