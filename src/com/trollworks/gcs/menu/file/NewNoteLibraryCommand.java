/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.notes.NotesDockable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

/** Provides the "New Equipment Library" command. */
public class NewNoteLibraryCommand extends Command {
	@Localize("New Note Library")
	@Localize(locale = "pt-br", value = "Nova biblioteca de notas")
	private static String TITLE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String					CMD_NEW_LIBRARY	= "NewNoteLibrary";				//$NON-NLS-1$

	/** The singleton {@link NewNoteLibraryCommand}. */
	public static final NewNoteLibraryCommand	INSTANCE		= new NewNoteLibraryCommand();

	private NewNoteLibraryCommand() {
		super(TITLE, CMD_NEW_LIBRARY);
	}

	@Override
	public void adjust() {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newNoteLibrary();
	}

	/** @return The newly created a new {@link NotesDockable}. */
	public static NotesDockable newNoteLibrary() {
		LibraryExplorerDockable library = LibraryExplorerDockable.get();
		if (library != null) {
			NoteList list = new NoteList();
			list.getModel().setLocked(false);
			NotesDockable dockable = new NotesDockable(list);
			library.dockLibrary(dockable);
			return dockable;
		}
		return null;
	}
}
