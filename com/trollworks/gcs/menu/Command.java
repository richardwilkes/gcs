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
import java.awt.KeyboardFocusManager;
import java.awt.Toolkit;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.InputEvent;

import javax.swing.AbstractAction;
import javax.swing.Icon;
import javax.swing.JCheckBoxMenuItem;
import javax.swing.JMenuItem;
import javax.swing.JRadioButtonMenuItem;
import javax.swing.KeyStroke;

/** Represents a command in the user interface. */
public abstract class Command extends AbstractAction {
	/** The standard command modifier for this platform. */
	public static final int	COMMAND_MODIFIER			= Toolkit.getDefaultToolkit().getMenuShortcutKeyMask();
	/** The standard command modifier for this platform, plus the shift key. */
	public static final int	SHIFTED_COMMAND_MODIFIER	= COMMAND_MODIFIER | InputEvent.SHIFT_DOWN_MASK;
	private boolean			mMarked;

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 */
	public Command(String title) {
		super(title);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 */
	public Command(String title, String command) {
		super(title);
		setCommand(command);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 */
	public Command(String title, Icon icon) {
		super(title, icon);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param icon The icon to use.
	 */
	public Command(String title, String command, Icon icon) {
		super(title, icon);
		setCommand(command);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param keyCode The key code to use. The platform's standard menu shortcut key will be
	 *            specified as a modifier.
	 */
	public Command(String title, int keyCode) {
		super(title);
		setAccelerator(keyCode);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param keyCode The key code to use. The platform's standard menu shortcut key will be
	 *            specified as a modifier.
	 */
	public Command(String title, String command, int keyCode) {
		super(title);
		setAccelerator(keyCode);
		setCommand(command);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param keyCode The key code to use.
	 * @param modifiers The modifiers to use.
	 */
	public Command(String title, int keyCode, int modifiers) {
		super(title);
		setAccelerator(keyCode, modifiers);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param keyCode The key code to use.
	 * @param modifiers The modifiers to use.
	 */
	public Command(String title, String command, int keyCode, int modifiers) {
		super(title);
		setAccelerator(keyCode, modifiers);
		setCommand(command);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 * @param keyCode The key code to use. The platform's standard menu shortcut key will be
	 *            specified as a modifier.
	 */
	public Command(String title, Icon icon, int keyCode) {
		super(title, icon);
		setAccelerator(keyCode);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param icon The icon to use.
	 * @param keyCode The key code to use. The platform's standard menu shortcut key will be
	 *            specified as a modifier.
	 */
	public Command(String title, String command, Icon icon, int keyCode) {
		super(title, icon);
		setAccelerator(keyCode);
		setCommand(command);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 * @param keyCode The key code to use.
	 * @param modifiers The modifiers to use.
	 */
	public Command(String title, Icon icon, int keyCode, int modifiers) {
		super(title, icon);
		setAccelerator(keyCode, modifiers);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param icon The icon to use.
	 * @param keyCode The key code to use.
	 * @param modifiers The modifiers to use.
	 */
	public Command(String title, String command, Icon icon, int keyCode, int modifiers) {
		super(title, icon);
		setAccelerator(keyCode, modifiers);
		setCommand(command);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param keystroke The {@link KeyStroke} to use.
	 */
	public Command(String title, KeyStroke keystroke) {
		super(title);
		setAccelerator(keystroke);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param keystroke The {@link KeyStroke} to use.
	 */
	public Command(String title, String command, KeyStroke keystroke) {
		super(title);
		setAccelerator(keystroke);
		setCommand(command);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 * @param keystroke The {@link KeyStroke} to use.
	 */
	public Command(String title, Icon icon, KeyStroke keystroke) {
		super(title, icon);
		setAccelerator(keystroke);
	}

	/**
	 * Creates a new {@link Command}.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param icon The icon to use.
	 * @param keystroke The {@link KeyStroke} to use.
	 */
	public Command(String title, String command, Icon icon, KeyStroke keystroke) {
		super(title, icon);
		setAccelerator(keystroke);
		setCommand(command);
	}

	/**
	 * Called to adjust the action prior to a menu being displayed.
	 * 
	 * @param item The {@link JMenuItem} that is using the {@link Command}. May be
	 *            <code>null</code>.
	 */
	public abstract void adjustForMenu(JMenuItem item);

	public abstract void actionPerformed(ActionEvent event);

	/** @return The {@link Command}'s title. */
	public final String getTitle() {
		Object value = getValue(NAME);
		return value != null ? value.toString() : null;
	}

	/** @param title The {@link Command}'s title. */
	public final void setTitle(String title) {
		putValue(NAME, title);
	}

	/** @return The {@link Command}'s command string. */
	public final String getCommand() {
		Object value = getValue(ACTION_COMMAND_KEY);
		return value != null ? value.toString() : null;
	}

	/** @param cmd The {@link Command}'s command string. */
	public final void setCommand(String cmd) {
		putValue(ACTION_COMMAND_KEY, cmd);
	}

	/** @return The current keyboard accelerator for this {@link Command}, or <code>null</code>. */
	public final KeyStroke getAccelerator() {
		Object value = getValue(ACCELERATOR_KEY);
		return value instanceof KeyStroke ? (KeyStroke) value : null;
	}

	/**
	 * Sets the keyboard accelerator for this {@link Command}. The platform's standard menu
	 * shortcut key will be specified as a modifier.
	 * 
	 * @param keyCode The key code to use.
	 */
	public final void setAccelerator(int keyCode) {
		setAccelerator(keyCode, COMMAND_MODIFIER);
	}

	/**
	 * Sets the keyboard accelerator for this {@link Command}.
	 * 
	 * @param keyCode The key code to use.
	 * @param modifiers The modifiers to use.
	 */
	public final void setAccelerator(int keyCode, int modifiers) {
		setAccelerator(KeyStroke.getKeyStroke(keyCode, modifiers));
	}

	/**
	 * Sets the keyboard accelerator for this {@link Command}.
	 * 
	 * @param keystroke The {@link KeyStroke} to use.
	 */
	public final void setAccelerator(KeyStroke keystroke) {
		DynamicMenuEnabler.remove(this);
		putValue(ACCELERATOR_KEY, keystroke);
		DynamicMenuEnabler.add(this);
	}

	/** Removes any previously set keyboard accelerator. */
	public final void removeAccelerator() {
		DynamicMenuEnabler.remove(this);
		putValue(ACCELERATOR_KEY, null);
	}

	/** @return Whether any associated menu item should be marked. */
	public final boolean isMarked() {
		return mMarked;
	}

	/** @param marked Whether any associated menu item should be marked. */
	public final void setMarked(boolean marked) {
		mMarked = marked;
	}

	/**
	 * @param item The {@link JMenuItem} to update to match the current marked state. May be
	 *            <code>null</code>.
	 */
	public final void updateMark(JMenuItem item) {
		if (item instanceof JCheckBoxMenuItem || item instanceof JRadioButtonMenuItem) {
			item.setSelected(mMarked);
		}
	}

	/** @return The current permanent focus owner. */
	public static final Component getFocusOwner() {
		return KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
	}

	/** @return The current active window. */
	public static final Window getActiveWindow() {
		return KeyboardFocusManager.getCurrentKeyboardFocusManager().getActiveWindow();
	}
}
