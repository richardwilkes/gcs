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

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.advantage.AdvantagesDockable;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Advantages Library" command. */
public class NewAdvantagesLibraryCommand extends Command {
	@Localize("New Advantages Library")
	private static String							TITLE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String						CMD_NEW_LIBRARY	= "NewAdvantagesLibrary";				//$NON-NLS-1$

	/** The singleton {@link NewAdvantagesLibraryCommand}. */
	public static final NewAdvantagesLibraryCommand	INSTANCE		= new NewAdvantagesLibraryCommand();

	private NewAdvantagesLibraryCommand() {
		super(TITLE, CMD_NEW_LIBRARY, KeyEvent.VK_L);
	}

	@Override
	public void adjust() {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newAdvantagesLibrary();
	}

	/** @return The newly created a new {@link AdvantagesDockable}. */
	public static AdvantagesDockable newAdvantagesLibrary() {
		LibraryExplorerDockable library = LibraryExplorerDockable.get();
		if (library != null) {
			AdvantagesDockable dockable = new AdvantagesDockable(new AdvantageList());
			library.dockLibrary(dockable);
			return dockable;
		}
		return null;
	}
}
