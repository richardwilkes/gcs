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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.common.ListWindow;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.awt.event.ActionEvent;

import javax.swing.JMenuItem;

/** Provides the "New List..." command. */
public class NewListCommand extends Command {
	private static String				MSG_NEW_ADVANTAGE_LIST;
	private static String				MSG_NEW_SKILL_LIST;
	private static String				MSG_NEW_SPELL_LIST;
	private static String				MSG_NEW_EQUIPMENT_LIST;

	static {
		LocalizedMessages.initialize(NewListCommand.class);
	}

	/** The new advantage list command. */
	public static final NewListCommand	ADVANTAGES	= new NewListCommand(MSG_NEW_ADVANTAGE_LIST, AdvantageList.class);
	/** The new advantage list command. */
	public static final NewListCommand	SKILLS		= new NewListCommand(MSG_NEW_SKILL_LIST, SkillList.class);
	/** The new advantage list command. */
	public static final NewListCommand	SPELLS		= new NewListCommand(MSG_NEW_SPELL_LIST, SpellList.class);
	/** The new advantage list command. */
	public static final NewListCommand	EQUIPMENT	= new NewListCommand(MSG_NEW_EQUIPMENT_LIST, EquipmentList.class);

	private Class<? extends ListFile>	mType;

	/**
	 * Creates a new {@link NewListCommand}.
	 * 
	 * @param title The title to use.
	 * @param type The type of list to create.
	 */
	public NewListCommand(String title, Class<? extends ListFile> type) {
		super(title);
		mType = type;
	}

	@Override public void adjustForMenu(JMenuItem item) {
		// Do nothing. We're always enabled.
	}

	@Override public void actionPerformed(ActionEvent event) {
		newList();
	}

	/** @return The newly created a new {@link ListWindow}. */
	public ListWindow newList() {
		try {
			return ListWindow.displayListWindow(mType.newInstance());
		} catch (Exception exception) {
			exception.printStackTrace(System.err);
			return null;
		}
	}
}
