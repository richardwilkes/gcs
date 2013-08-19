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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** The "Item" menu. */
public class ItemMenu extends JMenu {
	private static String	MSG_ITEM;

	static {
		LocalizedMessages.initialize(ItemMenu.class);
	}

	/** Creates a new {@link ItemMenu}. */
	public ItemMenu() {
		super(MSG_ITEM);
		add(new JMenuItem(OpenEditorCommand.INSTANCE));
		add(new JMenuItem(CopyToSheetCommand.INSTANCE));
		add(new JMenuItem(CopyToTemplateCommand.INSTANCE));
		add(new JMenuItem(ApplyTemplateCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(NewAdvantageCommand.INSTANCE));
		add(new JMenuItem(NewAdvantageCommand.CONTAINER_INSTANCE));
		addSeparator();
		add(new JMenuItem(NewSkillCommand.INSTANCE));
		add(new JMenuItem(NewSkillCommand.CONTAINER_INSTANCE));
		add(new JMenuItem(NewSkillCommand.TECHNIQUE));
		addSeparator();
		add(new JMenuItem(NewSpellCommand.INSTANCE));
		add(new JMenuItem(NewSpellCommand.CONTAINER_INSTANCE));
		addSeparator();
		add(new JMenuItem(NewEquipmentCommand.CARRIED_INSTANCE));
		add(new JMenuItem(NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE));
		DynamicMenuEnabler.add(this);
	}
}
