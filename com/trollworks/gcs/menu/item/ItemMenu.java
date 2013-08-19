/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2011 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.item;

import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.menu.DynamicMenuEnabler;
import com.trollworks.ttk.menu.DynamicMenuItem;
import com.trollworks.ttk.utility.LocalizedMessages;

import java.util.HashSet;

import javax.swing.JMenu;

/** The "Item" menu. */
public class ItemMenu extends JMenu {
	private static String	MSG_ITEM;

	static {
		LocalizedMessages.initialize(ItemMenu.class);
	}

	/**
	 * @return The set of {@link Command}s that this menu provides that can have their accelerators
	 *         modified.
	 */
	public static HashSet<Command> getCommands() {
		HashSet<Command> cmds = new HashSet<Command>();
		cmds.add(OpenEditorCommand.INSTANCE);
		cmds.add(CopyToSheetCommand.INSTANCE);
		cmds.add(CopyToTemplateCommand.INSTANCE);
		cmds.add(ApplyTemplateCommand.INSTANCE);
		cmds.add(NewAdvantageCommand.INSTANCE);
		cmds.add(NewAdvantageCommand.CONTAINER_INSTANCE);
		cmds.add(NewSkillCommand.INSTANCE);
		cmds.add(NewSkillCommand.CONTAINER_INSTANCE);
		cmds.add(NewSkillCommand.TECHNIQUE);
		cmds.add(NewSpellCommand.INSTANCE);
		cmds.add(NewSpellCommand.CONTAINER_INSTANCE);
		cmds.add(NewEquipmentCommand.CARRIED_INSTANCE);
		cmds.add(NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE);
		return cmds;
	}

	/** Creates a new {@link ItemMenu}. */
	public ItemMenu() {
		super(MSG_ITEM);
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
		add(new DynamicMenuItem(NewSkillCommand.TECHNIQUE));
		addSeparator();
		add(new DynamicMenuItem(NewSpellCommand.INSTANCE));
		add(new DynamicMenuItem(NewSpellCommand.CONTAINER_INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(NewEquipmentCommand.CARRIED_INSTANCE));
		add(new DynamicMenuItem(NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE));
		DynamicMenuEnabler.add(this);
	}
}
