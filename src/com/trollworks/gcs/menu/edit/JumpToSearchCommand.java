/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.edit;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;


import com.trollworks.toolkit.ui.menu.Command;

import java.awt.Component;
import java.awt.KeyboardFocusManager;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "Jump To Search" command. */
public class JumpToSearchCommand extends Command {
	@Localize("Jump To Search")
	private static String JUMP_TO_SEARCH;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String				CMD_JUMP_TO_SEARCH	= "JumpToSearch";				//$NON-NLS-1$

	/** The singleton {@link JumpToSearchCommand}. */
	public static final JumpToSearchCommand	INSTANCE			= new JumpToSearchCommand();

	private JumpToSearchCommand() {
		super(JUMP_TO_SEARCH, CMD_JUMP_TO_SEARCH, KeyEvent.VK_J);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		KeyboardFocusManager mgr = KeyboardFocusManager.getCurrentKeyboardFocusManager();
		Component focus = mgr.getPermanentFocusOwner();
		if (!(focus instanceof JumpToSearchTarget)) {
			focus = mgr.getActiveWindow();
		}
		setEnabled(focus instanceof JumpToSearchTarget);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
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
