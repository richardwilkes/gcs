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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Character Sheet" command. */
// RAW: Implement as dockables... need to support having a null backing file first
public class NewCharacterSheetCommand extends Command {
	@Localize("New Character Sheet")
	private static String							NEW_CHARACTER_SHEET;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String						CMD_NEW_CHARACTER_SHEET	= "NewCharacterSheet";				//$NON-NLS-1$

	/** The singleton {@link NewCharacterSheetCommand}. */
	public static final NewCharacterSheetCommand	INSTANCE				= new NewCharacterSheetCommand();

	private NewCharacterSheetCommand() {
		super(NEW_CHARACTER_SHEET, CMD_NEW_CHARACTER_SHEET, KeyEvent.VK_N);
	}

	@Override
	public void adjust() {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newSheet();
	}

	/** @return The newly created a new {@link SheetWindow}. */
	public static SheetWindow newSheet() {
		return SheetWindow.displaySheetWindow(new GURPSCharacter());
	}
}
