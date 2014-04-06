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


import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.DynamicMenuItem;

import java.util.HashSet;

import javax.swing.JMenu;

/** The "Item" menu. */
public class ItemMenu extends JMenu {
	@Localize("Item")
	private static String ITEM;

	static {
		Localization.initialize();
	}

	/**
	 * @return The set of {@link Command}s that this menu provides that can have their accelerators
	 *         modified.
	 */
	public static HashSet<Command> getCommands() {
		HashSet<Command> cmds = new HashSet<>();
		cmds.add(OpenEditorCommand.INSTANCE);
		cmds.add(CopyToSheetCommand.INSTANCE);
		cmds.add(CopyToTemplateCommand.INSTANCE);
		cmds.add(ApplyTemplateCommand.INSTANCE);
		cmds.add(NewAdvantageCommand.INSTANCE);
		cmds.add(NewAdvantageCommand.CONTAINER_INSTANCE);
		cmds.add(NewSkillCommand.INSTANCE);
		cmds.add(NewSkillCommand.CONTAINER_INSTANCE);
		cmds.add(NewSkillCommand.TECHNIQUE_INSTANCE);
		cmds.add(NewSpellCommand.INSTANCE);
		cmds.add(NewSpellCommand.CONTAINER_INSTANCE);
		cmds.add(NewEquipmentCommand.CARRIED_INSTANCE);
		cmds.add(NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE);
		return cmds;
	}

	/** Creates a new {@link ItemMenu}. */
	public ItemMenu() {
		super(ITEM);
		add(new DynamicMenuItem(OpenEditorCommand.INSTANCE));
		add(new DynamicMenuItem(CopyToSheetCommand.INSTANCE));
		add(new DynamicMenuItem(CopyToTemplateCommand.INSTANCE));
		add(new DynamicMenuItem(ApplyTemplateCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(NewAdvantageCommand.INSTANCE));
		add(new DynamicMenuItem(NewAdvantageCommand.CONTAINER_INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(NewSkillCommand.INSTANCE));
		add(new DynamicMenuItem(NewSkillCommand.CONTAINER_INSTANCE));
		add(new DynamicMenuItem(NewSkillCommand.TECHNIQUE_INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(NewSpellCommand.INSTANCE));
		add(new DynamicMenuItem(NewSpellCommand.CONTAINER_INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(NewEquipmentCommand.CARRIED_INSTANCE));
		add(new DynamicMenuItem(NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE));
		DynamicMenuEnabler.add(this);
	}
}
