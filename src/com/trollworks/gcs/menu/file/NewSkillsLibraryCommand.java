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
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.SkillsDockable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

/** Provides the "New Skills Library" command. */
public class NewSkillsLibraryCommand extends Command {
	@Localize("New Skills Library")
	@Localize(locale = "de", value = "Neue Fertigkeiten-Liste")
	@Localize(locale = "ru", value = "Новая библиотека умений")
	private static String						TITLE;

	static {
		Localization.initialize();
	}

	/** The action command this command will issue. */
	public static final String					CMD_NEW_LIBRARY	= "NewSkillsLibrary";				//$NON-NLS-1$

	/** The singleton {@link NewSkillsLibraryCommand}. */
	public static final NewSkillsLibraryCommand	INSTANCE		= new NewSkillsLibraryCommand();

	private NewSkillsLibraryCommand() {
		super(TITLE, CMD_NEW_LIBRARY);
	}

	@Override
	public void adjust() {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newSkillsLibrary();
	}

	/** @return The newly created a new {@link SkillsDockable}. */
	public static SkillsDockable newSkillsLibrary() {
		LibraryExplorerDockable library = LibraryExplorerDockable.get();
		if (library != null) {
			SkillList list = new SkillList();
			list.getModel().setLocked(false);
			SkillsDockable dockable = new SkillsDockable(list);
			library.dockLibrary(dockable);
			return dockable;
		}
		return null;
	}
}
