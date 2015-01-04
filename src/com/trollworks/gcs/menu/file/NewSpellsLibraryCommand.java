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

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.spell.SpellsDockable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

/** Provides the "New Spells Library" command. */
public class NewSpellsLibraryCommand extends Command {
	@Localize("New Spells Library")
	@Localize(locale = "de", value = "Neue Zauber-Liste")
	@Localize(locale = "ru", value = "Новая библиотека заклинаний")
	private static String						TITLE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String					CMD_NEW_LIBRARY	= "NewSpellsLibrary";				//$NON-NLS-1$

	/** The singleton {@link NewSpellsLibraryCommand}. */
	public static final NewSpellsLibraryCommand	INSTANCE		= new NewSpellsLibraryCommand();

	private NewSpellsLibraryCommand() {
		super(TITLE, CMD_NEW_LIBRARY);
	}

	@Override
	public void adjust() {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newSpellsLibrary();
	}

	/** @return The newly created a new {@link SpellsDockable}. */
	public static SpellsDockable newSpellsLibrary() {
		LibraryExplorerDockable library = LibraryExplorerDockable.get();
		if (library != null) {
			SpellList list = new SpellList();
			list.getModel().setLocked(false);
			SpellsDockable dockable = new SpellsDockable(list);
			library.dockLibrary(dockable);
			return dockable;
		}
		return null;
	}
}
