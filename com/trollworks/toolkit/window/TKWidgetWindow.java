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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.undo.TKUndoManager;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.menu.TKBaseMenu;
import com.trollworks.toolkit.widget.menu.TKMenuBar;

import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.Window;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;

/**
 * Provides the basis of a widget window, used to implement widgets that need to be able to extend
 * beyond the borders of their containing windows.
 */
public class TKWidgetWindow extends Window implements TKBaseWindow, ComponentListener {
	private TKPanel				mUserContent;
	private TKRepaintManager	mRepaintManager;

	/**
	 * Creates a new widget window.
	 * 
	 * @param owner The owning window.
	 */
	public TKWidgetWindow(TKBaseWindow owner) {
		super((Window) owner);
		setFocusableWindowState(false);
		setLayout(new TKCompassLayout());

		mRepaintManager = new TKRepaintManager(this);

		TKPanel content = new TKPanel(new TKCompassLayout());
		content.setOpaque(true);
		setContent(content);

		getOwner().addComponentListener(this);
	}

	@Override public void dispose() {
		getOwner().removeComponentListener(this);
		super.dispose();
	}

	public TKPanel getContent() {
		return mUserContent;
	}

	public void setContent(TKPanel panel) {
		if (mUserContent != null) {
			mUserContent.removeFromParent();
		}
		mUserContent = panel;
		add(panel);
	}

	private void closeMenus() {
		TKBaseMenu menu = ((TKBaseWindow) getOwner()).getUserInputManager().getMenuInUse();

		if (menu != null) {
			menu.closeCompletely(false);
		}
	}

	public void componentHidden(ComponentEvent event) {
		// Not used.
	}

	public void componentMoved(ComponentEvent event) {
		closeMenus();
	}

	public void componentResized(ComponentEvent event) {
		closeMenus();
	}

	public void componentShown(ComponentEvent event) {
		// Not used.
	}

	public TKRepaintManager getRepaintManager() {
		return mRepaintManager;
	}

	public TKUndoManager getUndoManager() {
		return ((TKBaseWindow) getOwner()).getUndoManager();
	}

	public TKUserInputManager getUserInputManager() {
		return ((TKBaseWindow) getOwner()).getUserInputManager();
	}

	public void forceFocusToAccept() {
		// Nothing to do
	}

	public Rectangle getLocalBounds(boolean insets) {
		Rectangle bounds = getBounds();

		bounds.x = 0;
		bounds.y = 0;
		if (insets) {
			Insets theInsets = getInsets();

			bounds.x += theInsets.left;
			bounds.y += theInsets.top;
			bounds.width -= theInsets.left + theInsets.right;
			bounds.height -= theInsets.top + theInsets.bottom;
		}
		return bounds;
	}

	public TKMenuBar getTKMenuBar() {
		return null;
	}

	public void setTKMenuBar(TKMenuBar menuBar) {
		// No menu bar allowed
	}

	public TKToolBar getTKToolBar() {
		return null;
	}

	public void setTKToolBar(TKToolBar toolBar) {
		// No tool bar allowed
	}

	public boolean isClosed() {
		return false;
	}

	public boolean isInForeground() {
		return false;
	}

	public boolean isPrinting() {
		return false;
	}

	public void processWindowKeyEvent(KeyEvent event) {
		// Not used by default.
	}

	public void processMouseEventSuper(MouseEvent event) {
		super.processMouseEvent(event);
	}

	public void processMouseMotionEventSuper(MouseEvent event) {
		super.processMouseMotionEvent(event);
	}
}
