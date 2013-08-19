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

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.awt.Window;
import java.awt.event.ActionEvent;

import javax.swing.JMenuItem;

/** Provides the "Add Natural Kick" command. */
public class AddNaturalKickCommand extends Command {
	private static String						MSG_ADD_NATURAL_KICK;

	static {
		LocalizedMessages.initialize(AddNaturalKickCommand.class);
	}

	/** The singleton {@link AddNaturalKickCommand}. */
	public static final AddNaturalKickCommand	INSTANCE	= new AddNaturalKickCommand();

	private AddNaturalKickCommand() {
		super(MSG_ADD_NATURAL_KICK);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof SheetWindow) {
			setEnabled(true);
			setMarked(((SheetWindow) window).getCharacter().includeKick());
		} else {
			setEnabled(false);
			setMarked(false);
		}
		updateMark(item);
	}

	@Override public void actionPerformed(ActionEvent event) {
		GURPSCharacter character = ((SheetWindow) getActiveWindow()).getCharacter();
		character.setIncludeKick(!character.includeKick());
	}
}
