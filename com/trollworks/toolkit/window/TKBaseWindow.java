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
import com.trollworks.toolkit.widget.menu.TKMenuBar;

import java.awt.Rectangle;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;

/** Methods that all OS-level windows must implement. */
public interface TKBaseWindow {
	/** @return The window's content panel. */
	public TKPanel getContent();

	/** @param content The window's content panel. */
	public void setContent(TKPanel content);

	/** @return The window's repaint manager. */
	public TKRepaintManager getRepaintManager();

	/** @return The window's undo manager. */
	public TKUndoManager getUndoManager();

	/** @return The window's user input manager. */
	public TKUserInputManager getUserInputManager();

	/**
	 * Call to force any existing text field with focus in the window to notify its action
	 * listeners.
	 */
	public void forceFocusToAccept();

	/**
	 * @param inset Pass in <code>true</code> to account for the window insets.
	 * @return The window's bounds in local coordinates.
	 */
	public Rectangle getLocalBounds(boolean inset);

	/** @return The window's menu bar. */
	public TKMenuBar getTKMenuBar();

	/** @param menuBar The window's menu bar. */
	public void setTKMenuBar(TKMenuBar menuBar);

	/** @return The window's tool bar. */
	public TKToolBar getTKToolBar();

	/** @param toolBar The window's tool bar. */
	public void setTKToolBar(TKToolBar toolBar);

	/** @return <code>true</code> if the window has been closed. */
	public boolean isClosed();

	/** @return <code>true</code> if the window is in the foreground. */
	public boolean isInForeground();

	/** @return <code>true</code> if the window is currently printing. */
	public boolean isPrinting();

	/**
	 * Calls <code>super.processMouseEvent(MouseEvent)</code>.
	 * 
	 * @param event The mouse event to process.
	 */
	public void processMouseEventSuper(MouseEvent event);

	/**
	 * Calls <code>super.processMouseMotionEvent(MouseEvent)</code>.
	 * 
	 * @param event The mouse event to process.
	 */
	public void processMouseMotionEventSuper(MouseEvent event);

	/**
	 * Called to process key events that weren't otherwise consumed.
	 * 
	 * @param event The key event.
	 */
	public void processWindowKeyEvent(KeyEvent event);
}
