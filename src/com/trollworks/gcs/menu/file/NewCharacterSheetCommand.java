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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Character Sheet" command. */
public class NewCharacterSheetCommand extends Command {
	@Localize("New Character Sheet")
	@Localize(locale = "de", value = "Neues Charakterblatt")
	@Localize(locale = "ru", value = "Новый лист персонажа")
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

	/** @return The newly created a new {@link SheetDockable}. */
	public static SheetDockable newSheet() {
		LibraryExplorerDockable library = LibraryExplorerDockable.get();
		if (library != null) {
			SheetDockable sheet = new SheetDockable(new GURPSCharacter());
			library.dockSheet(sheet);
			return sheet;
		}
		return null;
	}
}
