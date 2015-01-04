/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Jump To Search" command. */
public class JumpToSearchCommand extends Command {
	@Localize("Jump To Search")
	@Localize(locale = "de", value = "Springe zur Suche")
	@Localize(locale = "ru", value = "Перейти к поиску")
	private static String					JUMP_TO_SEARCH;

	static {
		Localization.initialize();
	}

	/** The singleton {@link JumpToSearchCommand}. */
	public static final JumpToSearchCommand	INSTANCE	= new JumpToSearchCommand();

	private JumpToSearchCommand() {
		super(JUMP_TO_SEARCH, null, KeyEvent.VK_J);
	}

	@Override
	public void adjust() {
		JumpToSearchTarget target = getTarget(JumpToSearchTarget.class);
		setEnabled(target != null && target.isJumpToSearchAvailable());
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		JumpToSearchTarget target = getTarget(JumpToSearchTarget.class);
		if (target != null) {
			target.jumpToSearchField();
		}
	}
}
