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

import java.awt.KeyEventDispatcher;
import java.awt.KeyEventPostProcessor;
import java.awt.KeyboardFocusManager;
import java.awt.Window;
import java.awt.event.KeyEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.util.BitSet;

import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.menu.TKMenuBar;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

/** A key dispatcher for dealing with menus and such. */
public class TKKeyDispatcher implements KeyEventDispatcher, KeyEventPostProcessor, PropertyChangeListener {
	private static BitSet	STATE	= new BitSet();

	/**
	 * @param keyCode The key code to check for.
	 * @return Whether the specified key code is currently pressed.
	 */
	public static boolean isKeyPressed(int keyCode) {
		return STATE.get(keyCode);
	}

	/** Creates a new key dispatcher. */
	public TKKeyDispatcher() {
		KeyboardFocusManager.getCurrentKeyboardFocusManager().addPropertyChangeListener("focusOwner", this); //$NON-NLS-1$
	}

	public boolean dispatchKeyEvent(KeyEvent event) {
		Window window = KeyboardFocusManager.getCurrentKeyboardFocusManager().getActiveWindow();

		if (event.getID() == KeyEvent.KEY_PRESSED) {
			STATE.set(event.getKeyCode());
		} else if (event.getID() == KeyEvent.KEY_RELEASED) {
			STATE.clear(event.getKeyCode());
		}
		TKUserInputManager.notifyMonitors(event);
		if (window instanceof TKBaseWindow) {
			TKBaseWindow baseWindow = (TKBaseWindow) window;
			TKMenuBar menuBar = baseWindow.getTKMenuBar();

			if (menuBar != null) {
				if (event.getID() == KeyEvent.KEY_PRESSED) {
					TKMenuItem item = menuBar.getMenuItemForKeyStroke(new TKKeystroke(event));

					if (item != null) {
						menuBar.menusWillBeAdjusted();
						if (!menuBar.adjustMenuItem(item.getCommand(), item)) {
							item.setEnabled(item.getSubMenu() != null);
						}
						menuBar.menusWereAdjusted();
						if (item.isEnabled()) {
							menuBar.issueCommand(item);
							event.consume();
							return true;
						}
					}
				}
			}
		}
		return false;
	}

	public boolean postProcessKeyEvent(KeyEvent event) {
		if (!event.isConsumed()) {
			Window window = KeyboardFocusManager.getCurrentKeyboardFocusManager().getActiveWindow();

			if (window instanceof TKBaseWindow) {
				TKBaseWindow baseWindow = (TKBaseWindow) window;

				baseWindow.processWindowKeyEvent(event);
				return event.isConsumed();
			}
		}
		return false;
	}

	public void propertyChange(PropertyChangeEvent event) {
		Object focus = event.getNewValue();

		if (focus instanceof TKPanel) {
			TKBaseWindow window = ((TKPanel) focus).getBaseWindow();

			if (window != null) {
				TKToolBar toolbar = window.getTKToolBar();

				if (toolbar != null) {
					toolbar.adjustToolBar();
				}
			}
		}
	}
}
