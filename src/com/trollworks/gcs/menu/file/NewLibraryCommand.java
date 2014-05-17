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

import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Library" command. */
//RAW: Implement as dockables... need to support having a null backing file first
public class NewLibraryCommand extends Command {
	@Localize("New Library")
	private static String					NEW_LIBRARY;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String				CMD_NEW_LIBRARY	= "NewLibrary";			//$NON-NLS-1$

	/** The singleton {@link NewLibraryCommand}. */
	public static final NewLibraryCommand	INSTANCE		= new NewLibraryCommand();

	private NewLibraryCommand() {
		super(NEW_LIBRARY, CMD_NEW_LIBRARY, KeyEvent.VK_L);
	}

	@Override
	public void adjust() {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newLibrary();
	}

	/** @return The newly created a new {@link LibraryWindow}. */
	public static LibraryWindow newLibrary() {
		return LibraryWindow.displayLibraryWindow(new LibraryFile());
	}
}
