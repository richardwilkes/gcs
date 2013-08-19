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

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.awt.Component;
import java.awt.Container;
import java.awt.Dialog;
import java.awt.Frame;
import java.awt.KeyboardFocusManager;
import java.awt.Window;
import java.awt.event.WindowAdapter;
import java.awt.event.WindowEvent;

import javax.swing.Icon;
import javax.swing.JDialog;
import javax.swing.JOptionPane;
import javax.swing.text.JTextComponent;

/** Utilities for use with windows. */
public class WindowUtils {
	private static String	MSG_ERROR;

	static {
		LocalizedMessages.initialize(WindowUtils.class);
	}

	/**
	 * @param comp The {@link Component} to use for determining the parent {@link Frame} or
	 *            {@link Dialog}.
	 * @param msg The message to display.
	 */
	public static void showError(Component comp, String msg) {
		JOptionPane.showMessageDialog(comp, msg, MSG_ERROR, JOptionPane.ERROR_MESSAGE);
	}

	/**
	 * Shows a confirmation dialog with custom options.
	 * 
	 * @param comp The {@link Component} to use. May be <code>null</code>.
	 * @param message The message.
	 * @param title The title to use.
	 * @param optionType The type of option dialog. Use the {@link JOptionPane} constants.
	 * @param options The options to display.
	 * @param initialValue The initial option.
	 * @return See the documentation for {@link JOptionPane}.
	 */
	public static int showConfirmDialog(Component comp, String message, String title, int optionType, Object[] options, Object initialValue) {
		return showOptionDialog(comp, message, title, false, optionType, JOptionPane.QUESTION_MESSAGE, null, options, initialValue);
	}

	/**
	 * Shows an option dialog.
	 * 
	 * @param parentComponent The parent {@link Component} to use. May be <code>null</code>.
	 * @param message The message. May be a {@link Component}.
	 * @param title The title to use.
	 * @param resizable Whether to allow the dialog to be resized by the user.
	 * @param optionType The type of option dialog. Use the {@link JOptionPane} constants.
	 * @param messageType The type of message. Use the {@link JOptionPane} constants.
	 * @param icon The icon to use. May be <code>null</code>.
	 * @param options The options to display. May be <code>null</code>.
	 * @param initialValue The initial option.
	 * @return See the documentation for {@link JOptionPane}.
	 */
	public static int showOptionDialog(Component parentComponent, Object message, String title, boolean resizable, int optionType, int messageType, Icon icon, Object[] options, Object initialValue) {
		JOptionPane pane = new JOptionPane(message, messageType, optionType, icon, options, initialValue);
		pane.setInitialValue(initialValue);
		pane.setComponentOrientation((parentComponent == null ? JOptionPane.getRootFrame() : parentComponent).getComponentOrientation());

		final JDialog dialog = pane.createDialog(parentComponent, title);
		new WindowSizeEnforcer(dialog);
		pane.selectInitialValue();
		dialog.setResizable(resizable);
		final Component field = getFirstFocusableField(message);
		if (field != null) {
			dialog.addWindowFocusListener(new WindowAdapter() {
				@Override public void windowGainedFocus(WindowEvent event) {
					field.requestFocus();
					dialog.removeWindowFocusListener(this);
				}
			});
		}
		dialog.setVisible(true);
		dialog.dispose();

		Object selectedValue = pane.getValue();
		if (selectedValue != null) {
			if (options == null) {
				if (selectedValue instanceof Integer) {
					return ((Integer) selectedValue).intValue();
				}
			} else {
				for (int i = 0; i < options.length; i++) {
					if (options[i].equals(selectedValue)) {
						return i;
					}
				}
			}
		}
		return JOptionPane.CLOSED_OPTION;
	}

	private static Component getFirstFocusableField(Object comp) {
		if (comp instanceof JTextComponent) {
			return (Component) comp;
		}
		if (comp instanceof Container) {
			for (Component child : ((Container) comp).getComponents()) {
				Component field = getFirstFocusableField(child);
				if (field != null) {
					return field;
				}
			}
		}
		return null;
	}

	/**
	 * @param comp The {@link Component} to use. May be <code>null</code>.
	 * @return The most logical {@link Window} associated with the component.
	 */
	public static Window getWindowForComponent(Component comp) {
		while (true) {
			if (comp == null) {
				return JOptionPane.getRootFrame();
			}
			if (comp instanceof Frame || comp instanceof Dialog) {
				return (Window) comp;
			}
			comp = comp.getParent();
		}
	}

	/**
	 * Temporarily removes focus and then restores it, forcing text fields to "commit" their
	 * contents.
	 */
	public static void forceFocusToAccept() {
		KeyboardFocusManager focusManager = KeyboardFocusManager.getCurrentKeyboardFocusManager();
		Component focus = focusManager.getPermanentFocusOwner();
		if (focus != null) {
			focusManager.clearGlobalFocusOwner();
			focus.requestFocus();
		}
	}
}
