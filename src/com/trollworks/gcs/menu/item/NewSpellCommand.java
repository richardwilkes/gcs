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

package com.trollworks.gcs.menu.item;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;


import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.toolkit.ui.menu.Command;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "New Spell" command. */
public class NewSpellCommand extends Command {
	@Localize("New Spell")
	private static String SPELL;
	@Localize("New Spell Container")
	private static String SPELL_CONTAINER;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String			CMD_SPELL			= "NewSpell";																								//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String			CMD_SPELL_CONTAINER	= "NewSpellContainer";																						//$NON-NLS-1$

	/** The "New Spell" command. */
	public static final NewSpellCommand	INSTANCE			= new NewSpellCommand(false, SPELL, CMD_SPELL, KeyEvent.VK_B, COMMAND_MODIFIER);
	/** The "New Spell Container" command. */
	public static final NewSpellCommand	CONTAINER_INSTANCE	= new NewSpellCommand(true, SPELL_CONTAINER, CMD_SPELL_CONTAINER, KeyEvent.VK_B, SHIFTED_COMMAND_MODIFIER);
	private boolean						mContainer;

	private NewSpellCommand(boolean container, String title, String cmd, int keyCode, int modifiers) {
		super(title, cmd, keyCode, modifiers);
		mContainer = container;
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			setEnabled(!((LibraryWindow) window).getOutline().getModel().isLocked());
		} else {
			setEnabled(window instanceof SheetWindow || window instanceof TemplateWindow);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		ListOutline outline;
		DataFile dataFile;

		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			LibraryWindow libraryWindow = (LibraryWindow) window;
			libraryWindow.switchToSpells();
			dataFile = libraryWindow.getLibraryFile();
			outline = libraryWindow.getOutline();
		} else if (window instanceof SheetWindow) {
			SheetWindow sheetWindow = (SheetWindow) window;
			outline = sheetWindow.getSheet().getSpellOutline();
			dataFile = sheetWindow.getCharacter();
		} else if (window instanceof TemplateWindow) {
			TemplateWindow templateWindow = (TemplateWindow) window;
			outline = templateWindow.getSheet().getSpellOutline();
			dataFile = templateWindow.getTemplate();
		} else {
			return;
		}

		Spell spell = new Spell(dataFile, mContainer);
		outline.addRow(spell, getTitle(), false);
		outline.getModel().select(spell, false);
		outline.scrollSelectionIntoView();
		outline.openDetailEditor(true);
	}
}
