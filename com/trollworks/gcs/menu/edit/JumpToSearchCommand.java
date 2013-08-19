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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.awt.Component;
import java.awt.KeyboardFocusManager;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "Jump To Search" command. */
public class JumpToSearchCommand extends Command {
	private static String					MSG_JUMP_TO_SEARCH;

	static {
		LocalizedMessages.initialize(JumpToSearchCommand.class);
	}

	/** The singleton {@link JumpToSearchCommand}. */
	public static final JumpToSearchCommand	INSTANCE	= new JumpToSearchCommand();

	private JumpToSearchCommand() {
		super(MSG_JUMP_TO_SEARCH, KeyEvent.VK_J);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		KeyboardFocusManager mgr = KeyboardFocusManager.getCurrentKeyboardFocusManager();
		Component focus = mgr.getPermanentFocusOwner();
		if (!(focus instanceof JumpToSearchTarget)) {
			focus = mgr.getActiveWindow();
		}
		setEnabled(focus instanceof JumpToSearchTarget);
	}

	@Override public void actionPerformed(ActionEvent event) {
		KeyboardFocusManager mgr = KeyboardFocusManager.getCurrentKeyboardFocusManager();
		Component focus = mgr.getPermanentFocusOwner();
		if (!(focus instanceof JumpToSearchTarget)) {
			focus = mgr.getActiveWindow();
		}
		if (focus instanceof JumpToSearchTarget) {
			((JumpToSearchTarget) focus).jumpToSearchField();
		}
	}
}
