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

package com.trollworks.gcs.menu.edit;

import static com.trollworks.gcs.menu.edit.AddNaturalPunchCommand_LS.*;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.menu.Command;

import java.awt.Window;
import java.awt.event.ActionEvent;

import javax.swing.JMenuItem;

@Localized({
				@LS(key = "ADD_NATURAL_PUNCH", msg = "Include Punch In Weapons"),
})
/** Provides the "Add Natural Punch" command. */
public class AddNaturalPunchCommand extends Command {
	/** The action command this command will issue. */
	public static final String					CMD_ADD_NATURAL_PUNCH	= "AddNaturalPunch";			//$NON-NLS-1$

	/** The singleton {@link AddNaturalPunchCommand}. */
	public static final AddNaturalPunchCommand	INSTANCE				= new AddNaturalPunchCommand();

	private AddNaturalPunchCommand() {
		super(ADD_NATURAL_PUNCH, CMD_ADD_NATURAL_PUNCH);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof SheetWindow) {
			setEnabled(true);
			setMarked(((SheetWindow) window).getCharacter().includePunch());
		} else {
			setEnabled(false);
			setMarked(false);
		}
		updateMark(item);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		GURPSCharacter character = ((SheetWindow) getActiveWindow()).getCharacter();
		character.setIncludePunch(!character.includePunch());
	}
}
