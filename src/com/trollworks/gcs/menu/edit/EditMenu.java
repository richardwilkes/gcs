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
import com.trollworks.toolkit.ui.menu.DynamicCheckBoxMenuItem;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.DynamicMenuItem;
import com.trollworks.toolkit.ui.menu.edit.CopyCommand;
import com.trollworks.toolkit.ui.menu.edit.CutCommand;
import com.trollworks.toolkit.ui.menu.edit.DeleteCommand;
import com.trollworks.toolkit.ui.menu.edit.PasteCommand;
import com.trollworks.toolkit.ui.menu.edit.PreferencesCommand;
import com.trollworks.toolkit.ui.menu.edit.RedoCommand;
import com.trollworks.toolkit.ui.menu.edit.SelectAllCommand;
import com.trollworks.toolkit.ui.menu.edit.UndoCommand;
import com.trollworks.toolkit.utility.Platform;

import java.util.HashSet;

import javax.swing.JMenu;

/** The standard "Edit" menu. */
public class EditMenu extends JMenu {
	@Localize("Edit")
	private static String EDIT;

	static {
		Localization.initialize();
	}

	/**
	 * @return The set of {@link Command}s that this menu provides that can have their accelerators
	 *         modified.
	 */
	public static HashSet<Command> getCommands() {
		HashSet<Command> cmds = new HashSet<>();
		cmds.add(UndoCommand.INSTANCE);
		cmds.add(RedoCommand.INSTANCE);
		cmds.add(CutCommand.INSTANCE);
		cmds.add(CopyCommand.INSTANCE);
		cmds.add(PasteCommand.INSTANCE);
		cmds.add(DuplicateCommand.INSTANCE);
		cmds.add(DeleteCommand.INSTANCE);
		cmds.add(SelectAllCommand.INSTANCE);
		cmds.add(IncrementCommand.INSTANCE);
		cmds.add(DecrementCommand.INSTANCE);
		cmds.add(RotateEquipmentStateCommand.INSTANCE);
		cmds.add(JumpToSearchCommand.INSTANCE);
		cmds.add(RandomizeDescriptionCommand.INSTANCE);
		cmds.add(RandomizeNameCommand.FEMALE_INSTANCE);
		cmds.add(RandomizeNameCommand.MALE_INSTANCE);
		cmds.add(AddNaturalPunchCommand.INSTANCE);
		cmds.add(AddNaturalKickCommand.INSTANCE);
		cmds.add(AddNaturalKickWithBootsCommand.INSTANCE);
		if (!Platform.isMacintosh()) {
			cmds.add(PreferencesCommand.INSTANCE);
		}
		return cmds;
	}

	/** Creates a new {@link EditMenu}. */
	public EditMenu() {
		super(EDIT);
		add(new DynamicMenuItem(UndoCommand.INSTANCE));
		add(new DynamicMenuItem(RedoCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(CutCommand.INSTANCE));
		add(new DynamicMenuItem(CopyCommand.INSTANCE));
		add(new DynamicMenuItem(PasteCommand.INSTANCE));
		add(new DynamicMenuItem(DuplicateCommand.INSTANCE));
		add(new DynamicMenuItem(DeleteCommand.INSTANCE));
		add(new DynamicMenuItem(SelectAllCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(IncrementCommand.INSTANCE));
		add(new DynamicMenuItem(DecrementCommand.INSTANCE));
		add(new DynamicMenuItem(RotateEquipmentStateCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(JumpToSearchCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(RandomizeDescriptionCommand.INSTANCE));
		add(new DynamicMenuItem(RandomizeNameCommand.FEMALE_INSTANCE));
		add(new DynamicMenuItem(RandomizeNameCommand.MALE_INSTANCE));
		addSeparator();
		add(new DynamicCheckBoxMenuItem(AddNaturalPunchCommand.INSTANCE));
		add(new DynamicCheckBoxMenuItem(AddNaturalKickCommand.INSTANCE));
		add(new DynamicCheckBoxMenuItem(AddNaturalKickWithBootsCommand.INSTANCE));
		if (!Platform.isMacintosh()) {
			addSeparator();
			add(new DynamicMenuItem(PreferencesCommand.INSTANCE));
		}
		DynamicMenuEnabler.add(this);
	}
}
