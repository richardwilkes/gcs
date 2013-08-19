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

package com.trollworks.gcs.menu;

import java.awt.Component;
import java.awt.KeyEventDispatcher;
import java.awt.KeyboardFocusManager;
import java.awt.event.KeyEvent;
import java.util.HashMap;

import javax.swing.Action;
import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.KeyStroke;
import javax.swing.event.MenuEvent;
import javax.swing.event.MenuListener;

/** Dynamically enables menu items just prior to them being used. */
public class DynamicMenuEnabler implements KeyEventDispatcher, MenuListener {
	private static final HashMap<KeyStroke, Command>	MAP			= new HashMap<KeyStroke, Command>();
	private static final DynamicMenuEnabler				INSTANCE	= new DynamicMenuEnabler();

	static {
		KeyboardFocusManager.getCurrentKeyboardFocusManager().addKeyEventDispatcher(INSTANCE);
	}

	private DynamicMenuEnabler() {
		// Just here to prevent external instantiation
	}

	/** @param cmd The {@link Command} to add. */
	public static void add(Command cmd) {
		MAP.put(cmd.getAccelerator(), cmd);
	}

	/** @param cmd The {@link Command} to remove. */
	public static void remove(Command cmd) {
		MAP.remove(cmd.getAccelerator());
	}

	/** @param menu The {@link JMenu} to add. */
	public static void add(JMenu menu) {
		menu.addMenuListener(INSTANCE);
	}

	/** @param menu The {@link JMenu} to remove. */
	public static void remove(JMenu menu) {
		menu.removeMenuListener(INSTANCE);
	}

	public boolean dispatchKeyEvent(KeyEvent event) {
		Command action = MAP.get(KeyStroke.getKeyStrokeForEvent(event));
		if (action != null) {
			action.adjustForMenu(null);
		}
		return false;
	}

	public void menuCanceled(MenuEvent event) {
		// Not used.
	}

	public void menuDeselected(MenuEvent event) {
		// Not used.
	}

	public void menuSelected(MenuEvent event) {
		JMenu menu = (JMenu) event.getSource();
		for (Component component : menu.getMenuComponents()) {
			if (component instanceof JMenuItem) {
				JMenuItem item = (JMenuItem) component;
				Action action = item.getAction();
				if (action instanceof Command) {
					((Command) action).adjustForMenu(item);
				}
			}
		}
	}
}
