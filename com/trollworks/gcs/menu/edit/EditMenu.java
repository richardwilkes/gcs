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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.edit;

import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.menu.DynamicCheckBoxMenuItem;
import com.trollworks.ttk.menu.DynamicMenuEnabler;
import com.trollworks.ttk.menu.DynamicMenuItem;
import com.trollworks.ttk.menu.edit.CopyCommand;
import com.trollworks.ttk.menu.edit.CutCommand;
import com.trollworks.ttk.menu.edit.DeleteCommand;
import com.trollworks.ttk.menu.edit.PasteCommand;
import com.trollworks.ttk.menu.edit.PreferencesCommand;
import com.trollworks.ttk.menu.edit.RedoCommand;
import com.trollworks.ttk.menu.edit.SelectAllCommand;
import com.trollworks.ttk.menu.edit.UndoCommand;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.Platform;

import java.util.HashSet;

import javax.swing.JMenu;

/** The standard "Edit" menu. */
public class EditMenu extends JMenu {
	private static String	MSG_EDIT;

	static {
		LocalizedMessages.initialize(EditMenu.class);
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
		super(MSG_EDIT);
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
