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

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.text.TKDocument;
import com.trollworks.toolkit.text.TKDocumentListener;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKPlatform;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.menu.TKBaseMenu;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.window.TKCorrectableManager;
import com.trollworks.toolkit.window.TKUserInputManager;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.LayoutManager2;
import java.awt.Rectangle;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;

/** A standard popup menu. */
public class TKPopupMenu extends TKBaseMenu implements TKMenuTarget, TKDocumentListener, LayoutManager2 {
	private TKMenu			mMenu;
	private TKMenuTarget	mTarget;
	private int				mSelection;
	private TKTextField		mEditField;
	private boolean			mExtra;
	private int				mPopupIconWidth;
	private int				mPopupIconHeight;
	private boolean			mInRollOver;
	private boolean			mBlankWhenDisabled;

	/**
	 * Creates a popup menu.
	 * 
	 * @param menu The menu to be displayed.
	 */
	public TKPopupMenu(TKMenu menu) {
		this(menu, null, false, 0);
	}

	/**
	 * Creates a popup menu.
	 * 
	 * @param menu The menu to be displayed.
	 * @param editable Pass in <code>true</code> if the popup should have a user-editable field.
	 */
	public TKPopupMenu(TKMenu menu, boolean editable) {
		this(menu, null, editable, 0);
	}

	/**
	 * Creates a popup menu.
	 * 
	 * @param menu The menu to be displayed.
	 * @param selection The item index to select within the menu.
	 */
	public TKPopupMenu(TKMenu menu, int selection) {
		this(menu, null, false, selection);
	}

	/**
	 * Creates a popup menu.
	 * 
	 * @param menu The menu to be displayed.
	 * @param target The menu target. Pass in <code>null</code> if the popup menu should deal with
	 *            menu selections on its own.
	 * @param editable Pass in <code>true</code> if the popup should have a user-editable field.
	 */
	public TKPopupMenu(TKMenu menu, TKMenuTarget target, boolean editable) {
		this(menu, target, editable, 0);
	}

	/**
	 * Creates a popup menu.
	 * 
	 * @param menu The menu to be displayed.
	 * @param target The menu target. Pass in <code>null</code> if the popup menu should deal with
	 *            menu selections on its own.
	 * @param editable Pass in <code>true</code> if the popup should have a user-editable field.
	 * @param selection The item index to select within the menu.
	 */
	public TKPopupMenu(TKMenu menu, TKMenuTarget target, boolean editable, int selection) {
		super();

		BufferedImage icon = TKImage.getDownArrowIcon();
		mPopupIconWidth = icon.getWidth();
		mPopupIconHeight = icon.getHeight();

		mMenu = menu;
		mTarget = target;
		if (selection < 0) {
			selection = 0;
		} else {
			int menuItemCount = menu.getMenuItemCount();

			if (selection >= menuItemCount) {
				selection = menuItemCount - 1;
			}
		}
		setOpaque(true);
		setCursor(Cursor.getDefaultCursor());
		setLayout(this);
		if (editable) {
			mEditField = new EditField();
			add(mEditField);
		}
		setSelectedItem(selection);
	}

	/** @return The underlying menu. */
	public TKMenu getMenu() {
		return mMenu;
	}

	/**
	 * Replaces the current menu with the specified menu.
	 * 
	 * @param menu The menu to set.
	 * @param indexToSelect The index to select within the menu.
	 */
	public void setMenu(TKMenu menu, int indexToSelect) {
		mMenu = menu;
		setSelectedItem(indexToSelect);
	}

	public void addLayoutComponent(String name, Component component) {
		// Nothing to do...
	}

	public void addLayoutComponent(Component component, Object constraints) {
		// Nothing to do...
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		return mTarget != null ? mTarget.adjustMenuItem(command, item) : true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		boolean processed = mTarget != null ? mTarget.obeyCommand(command, item) : true;

		setSelectedItem(mMenu.getMenuItemIndex(item));

		return processed;
	}

	@Override public void close(boolean commandWillBeProcessed) {
		mMenu.close(commandWillBeProcessed);
		super.close(commandWillBeProcessed);
		if (mEditField != null) {
			mEditField.requestFocus();
		} else {
			TKCorrectableManager.getInstance().clearAllCorrectables();
		}
	}

	public void documentChanged(TKDocument document) {
		String text = mEditField.getText();
		int length = getMenuItemCount();

		mSelection = TKMenu.NOTHING_HIT;

		for (int i = 0; i < length - (mExtra ? 2 : 0); i++) {
			TKMenuItem item = getMenuItem(i);

			if (text.equals(item.getTitle())) {
				mSelection = i;
				break;
			}
		}

		if (mExtra) {
			mMenu.remove(mMenu.getMenuItem(--length));
			mMenu.remove(mMenu.getMenuItem(--length));
			mExtra = false;
		}

		if (mSelection == TKMenu.NOTHING_HIT) {
			mMenu.addSeparator();
			mMenu.add(new TKMenuItem(text));
			mSelection = length + 1;
			mExtra = true;
		}

		notifyActionListeners();
	}

	public float getLayoutAlignmentX(Container target) {
		return CENTER_ALIGNMENT;
	}

	public float getLayoutAlignmentY(Container target) {
		return CENTER_ALIGNMENT;
	}

	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = getPreferredSizeSelf();

		size.width = MAX_SIZE;
		return size;
	}

	/**
	 * @param index The item index.
	 * @return The menu item at the specified index.
	 */
	@Override public TKMenuItem getMenuItem(int index) {
		return mMenu.getMenuItem(index);
	}

	/**
	 * @param obj The item user object.
	 * @return The menu item with the specified user object.
	 */
	public TKMenuItem getMenuItemForUserObject(Object obj) {
		int count = getMenuItemCount();

		for (int i = 0; i < count; i++) {
			if (obj.equals(getMenuItem(i).getUserObject())) {
				return getMenuItem(i);
			}
		}
		return null;
	}

	/**
	 * @param obj The item user object.
	 * @return The menu item index with the specified user object, or -1.
	 */
	public int getMenuItemIndexForUserObject(Object obj) {
		int count = getMenuItemCount();

		for (int i = 0; i < count; i++) {
			if (obj.equals(getMenuItem(i).getUserObject())) {
				return i;
			}
		}
		return -1;
	}

	/**
	 * @param title The menu item title.
	 * @return The menu item index with the specified title, or -1.
	 */
	public int getMenuItemIndexForTitle(String title) {
		int count = getMenuItemCount();

		for (int i = 0; i < count; i++) {
			if (title.equals(getMenuItem(i).getTitle())) {
				return i;
			}
		}
		return -1;
	}

	/** @return The number of menu items in this popup. */
	@Override public int getMenuItemCount() {
		return mMenu.getMenuItemCount();
	}

	@Override protected Dimension getMinimumSizeSelf() {
		return getPreferredSizeSelf();
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Insets insets = getInsets();
		int count = getMenuItemCount();
		int height = 0;
		int width = 0;

		for (int i = 0; i < count; i++) {
			TKMenuItem item = getMenuItem(i);
			int tmp;

			item.setFullDisplay(false);
			tmp = item.getWidth(item.getPreferredDividingPoint());
			if (tmp > width) {
				width = tmp;
			}
			tmp = item.getHeight();
			if (tmp > height) {
				height = tmp - 1;
			}
			item.setFullDisplay(true);
			mMenu.getPreferredSize(); // Needed to cause the menu to restore its cached sizes
		}

		if (height < mPopupIconHeight) {
			height = mPopupIconHeight;
		}

		if (mEditField != null) {
			Dimension size = mEditField.getPreferredSize();

			if (width < size.width) {
				width = size.width;
			}
			if (height < size.height) {
				height = size.height;
			}
		}

		return new Dimension(insets.left + insets.right + width + mPopupIconWidth + TKMenuItem.GAP, insets.top + insets.bottom + height);
	}

	/** @return The selected menu item. */
	public TKMenuItem getSelectedItem() {
		return getMenuItem(mSelection);
	}

	/** @return The selected menu item index. */
	public int getSelectedItemIndex() {
		return mSelection;
	}

	/** @return The selected menu item user object. */
	public Object getSelectedItemUserObject() {
		return getSelectedItem().getUserObject();
	}

	public void invalidateLayout(Container target) {
		// Nothing to do...
	}

	public void layoutContainer(Container target) {
		if (mEditField != null) {
			Insets insets = getInsets();

			mEditField.setBounds(insets.left, insets.top, getWidth() - (insets.left + insets.right + mPopupIconWidth + TKMenuItem.GAP), getHeight() - (insets.top + insets.bottom));
		}
	}

	public Dimension maximumLayoutSize(Container target) {
		return getMaximumSizeSelf();
	}

	public Dimension minimumLayoutSize(Container target) {
		return getMinimumSizeSelf();
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		TKMenuItem item = getMenuItem(mSelection);
		Insets insets = getInsets();
		int height = getHeight();
		int width = getWidth();
		BufferedImage icon = TKImage.getDownArrowIcon();

		super.paintPanel(g2d, clips);

		g2d.setPaint(mInRollOver ? TKColor.CONTROL_ROLL : TKColor.CONTROL_FILL);
		g2d.fillRect(insets.left, insets.top, width, height);
		g2d.setPaint(TKColor.CONTROL_LINE);
		g2d.drawRect(insets.left, insets.top, width - 1, height - 1);

		g2d.setPaint(TKColor.CONTROL_SHADOW);
		int x = insets.left + width - 2;
		int y = insets.top + height - 2;
		g2d.drawLine(x, insets.top + 1, x, y);
		g2d.drawLine(insets.left + 1, y, x, y);

		g2d.setPaint(TKColor.CONTROL_HIGHLIGHT);
		g2d.drawLine(insets.left + 1, insets.top + 1, insets.left + 1, y);
		g2d.drawLine(insets.left + 1, insets.top + 1, x, insets.top + 1);

		x = width - (mPopupIconWidth + insets.right + 7);
		y = 1;

		g2d.setPaint(TKColor.CONTROL_SHADOW);
		g2d.drawLine(x, y, x, y + height - 3);
		g2d.setPaint(TKColor.CONTROL_HIGHLIGHT);
		g2d.drawLine(x + 1, y, x + 1, y + height - 3);

		if (mEditField == null && item != null) {
			boolean enabled = isEnabled();

			if (enabled || !isDisplayBlankWhenDisabled()) {
				boolean itemEnabled = item.isEnabled();
				int indent = item.getIndent();

				if (!enabled && itemEnabled) {
					item.setEnabled(false);
				}
				item.setIndent(0);
				item.setFullDisplay(false);
				item.getWidth(item.getPreferredDividingPoint()); // Called to ensure the next
				// draw call will draw
				// properly...
				item.draw(g2d, insets.left, insets.top, width - (insets.left + insets.right + TKMenuItem.GAP + mPopupIconWidth), height - (insets.top + insets.bottom), null);
				if (!isEnabled() && itemEnabled) {
					item.setEnabled(true);
				}
				item.setIndent(indent);
				item.setFullDisplay(true);
				mMenu.getPreferredSize(); // Needed to cause the menu to restore its cached sizes
			}
		}

		if (isEnabled()) {
			g2d.drawImage(icon, width - (mPopupIconWidth + insets.right + 3), (height - mPopupIconHeight) / 2, null);
		} else {
			g2d.drawImage(TKImage.createDisabledImage(icon), width - (mPopupIconWidth + insets.right + 3), (height - mPopupIconHeight) / 2, null);
		}
	}

	/** @return Whether the popup should be blank when disabled. */
	public boolean isDisplayBlankWhenDisabled() {
		return mBlankWhenDisabled;
	}

	/** @param blankWhenDisabled Whether the popup should be blank when disabled. */
	public void setDisplayBlankWhenDisabled(boolean blankWhenDisabled) {
		mBlankWhenDisabled = blankWhenDisabled;
	}

	public Dimension preferredLayoutSize(Container target) {
		return getPreferredSizeSelf();
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		switch (event.getID()) {
			case MouseEvent.MOUSE_PRESSED:
				int count = getMenuItemCount();

				for (int i = 0; i < count; i++) {
					getMenuItem(i).setMarked(i == mSelection);
				}
				mMenu.display(this, this, getLocalBounds(), mSelection);
				mInRollOver = false;
				repaint();
				break;
			case MouseEvent.MOUSE_RELEASED:
				if (mMenu.isOpen()) {
					TKUserInputManager.forwardMouseEvent(event, this, mMenu);
				}
				break;
			case MouseEvent.MOUSE_MOVED:
			case MouseEvent.MOUSE_DRAGGED:
				if (mMenu.isOpen()) {
					TKUserInputManager.forwardMouseEvent(event, this, mMenu);
				}
				break;
			case MouseEvent.MOUSE_ENTERED:
				if (!mMenu.isOpen()) {
					mInRollOver = true;
					repaint();
				}
				break;
			case MouseEvent.MOUSE_EXITED:
				mInRollOver = false;
				repaint();
				break;
		}
	}

	public void removeLayoutComponent(Component target) {
		// Nothing to do...
	}

	@Override public void setEnabled(boolean enabled) {
		if (enabled != isEnabled()) {
			if (mEditField != null) {
				mEditField.setEnabled(enabled);
			}
		}
		super.setEnabled(enabled);
	}

	/** @param item The menu item to select. */
	public void setSelectedItem(TKMenuItem item) {
		setSelectedItem(mMenu.getMenuItemIndex(item));
	}

	/** @param index The menu item index to select. */
	public void setSelectedItem(int index) {
		mSelection = index;
		if (mEditField != null) {
			TKMenuItem item = getMenuItem(index);

			if (item != null) {
				String title = item.getTitle();

				if (title == null) {
					title = ""; //$NON-NLS-1$
				}
				mEditField.setText(title);
				mEditField.selectAll();
				return;
			}
		}
		notifyActionListeners();
		repaint();
	}

	/** @param obj The user object to select. */
	public void setSelectedUserObject(Object obj) {
		if (obj != null) {
			int count = getMenuItemCount();

			for (int i = 0; i < count; i++) {
				if (obj.equals(getMenuItem(i).getUserObject())) {
					setSelectedItem(i);
					break;
				}
			}
		}
	}

	/**
	 * Searches the popup menu for a matching title.
	 * 
	 * @param title The title to look for.
	 * @return The index of the matching item, or {@link TKMenu#NOTHING_HIT} if there are no
	 *         matches.
	 */
	public int getMatchingItemIndex(String title) {
		if (title != null && title.length() > 0) {
			int length = getMenuItemCount();

			for (int i = 0; i < length; i++) {
				TKMenuItem item = getMenuItem(i);

				if (title.equals(item.getTitle())) {
					return i;
				}
			}
		}
		return TKMenu.NOTHING_HIT;
	}

	/**
	 * Sets the selected menu item by searching for the specified string. If this popup is editable
	 * and the string is not found, then the editable field will still be set appropriately.
	 * 
	 * @param title The title to search for.
	 */
	public void setSelectedItem(String title) {
		if (title != null && title.length() > 0) {
			if (mEditField != null) {
				mEditField.setText(title);
				mEditField.selectAll();
			} else {
				mSelection = getMatchingItemIndex(title);
				notifyActionListeners();
				repaint();
			}
		}
	}

	/**
	 * @return The edit field for an editable pop-up menu. If this is a non-editable pop-up, returns
	 *         <code>null</code>.
	 */
	public TKTextField getEditField() {
		return mEditField;
	}

	/**
	 * @param event The mouse event.
	 * @return <code>true</code> if this event is the platform's notion of a popup trigger.
	 */
	public static final boolean isPopupTrigger(MouseEvent event) {
		int button = event.getButton();

		return button == MouseEvent.BUTTON2 || button == MouseEvent.BUTTON3 || TKPlatform.isMacintosh() && event.isControlDown() && button != MouseEvent.NOBUTTON;
	}

	private class EditField extends TKTextField {
		/** Creates a new edit field for the popup control. */
		EditField() {
			super();
			setBorder(new TKCompoundBorder(new TKLineBorder(Color.darkGray), new TKEmptyBorder(2, 2, 1, 2)));
			addDocumentListener(TKPopupMenu.this);
		}

		@Override public String getToolTipText() {
			return TKPopupMenu.this.getToolTipText();
		}
	}
}
