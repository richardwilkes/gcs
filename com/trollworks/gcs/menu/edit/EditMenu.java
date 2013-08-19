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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import javax.swing.JCheckBoxMenuItem;
import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** The standard "Edit" menu. */
public class EditMenu extends JMenu {
	private static String	MSG_EDIT;

	static {
		LocalizedMessages.initialize(EditMenu.class);
	}

	/** Creates a new {@link EditMenu}. */
	public EditMenu() {
		super(MSG_EDIT);
		add(new JMenuItem(UndoCommand.INSTANCE));
		add(new JMenuItem(RedoCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(CutCommand.INSTANCE));
		add(new JMenuItem(CopyCommand.INSTANCE));
		add(new JMenuItem(PasteCommand.INSTANCE));
		add(new JMenuItem(DuplicateCommand.INSTANCE));
		add(new JMenuItem(DeleteCommand.INSTANCE));
		add(new JMenuItem(SelectAllCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(IncrementCommand.INSTANCE));
		add(new JMenuItem(DecrementCommand.INSTANCE));
		add(new JMenuItem(ToggleEquippedCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(JumpToSearchCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(RandomizeDescriptionCommand.INSTANCE));
		add(new JMenuItem(RandomizeNameCommand.FEMALE_INSTANCE));
		add(new JMenuItem(RandomizeNameCommand.MALE_INSTANCE));
		addSeparator();
		add(new JCheckBoxMenuItem(AddNaturalPunchCommand.INSTANCE));
		add(new JCheckBoxMenuItem(AddNaturalKickCommand.INSTANCE));
		add(new JCheckBoxMenuItem(AddNaturalKickWithBootsCommand.INSTANCE));
		if (!Platform.isMacintosh()) {
			addSeparator();
			add(new JMenuItem(PreferencesCommand.INSTANCE));
		}
		DynamicMenuEnabler.add(this);
	}
}
