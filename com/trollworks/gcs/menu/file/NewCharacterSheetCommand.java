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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.file;

import static com.trollworks.gcs.menu.file.NewCharacterSheetCommand_LS.*;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.menu.Command;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

@Localized({
				@LS(key = "NEW_CHARACTER_SHEET", msg = "New Character Sheet"),
})
/** Provides the "New Character Sheet" command. */
public class NewCharacterSheetCommand extends Command {
	/** The action command this command will issue. */
	public static final String						CMD_NEW_CHARACTER_SHEET	= "NewCharacterSheet";				//$NON-NLS-1$

	/** The singleton {@link NewCharacterSheetCommand}. */
	public static final NewCharacterSheetCommand	INSTANCE				= new NewCharacterSheetCommand();

	private NewCharacterSheetCommand() {
		super(NEW_CHARACTER_SHEET, CMD_NEW_CHARACTER_SHEET, KeyEvent.VK_N);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newSheet();
	}

	/** @return The newly created a new {@link SheetWindow}. */
	public static SheetWindow newSheet() {
		return SheetWindow.displaySheetWindow(new GURPSCharacter());
	}
}
