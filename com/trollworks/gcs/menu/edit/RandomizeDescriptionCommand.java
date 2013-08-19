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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.character.DescriptionRandomizer;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.WindowUtils;

import java.awt.Window;
import java.awt.event.ActionEvent;

import javax.swing.ImageIcon;
import javax.swing.JMenuItem;
import javax.swing.JOptionPane;

/** Provides the "Randomize Description" command. */
public class RandomizeDescriptionCommand extends Command {
	private static String							MSG_RANDOMIZE_DESCRIPTION;
	private static String							MSG_RANDOMIZER;
	private static String							MSG_APPLY;
	private static String							MSG_CANCEL;

	static {
		LocalizedMessages.initialize(RandomizeDescriptionCommand.class);
	}

	/** The singleton {@link RandomizeDescriptionCommand}. */
	public static final RandomizeDescriptionCommand	INSTANCE	= new RandomizeDescriptionCommand();

	private RandomizeDescriptionCommand() {
		super(MSG_RANDOMIZE_DESCRIPTION);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		setEnabled(window instanceof SheetWindow);
	}

	@Override public void actionPerformed(ActionEvent event) {
		DescriptionRandomizer panel = new DescriptionRandomizer(((SheetWindow) getActiveWindow()).getCharacter());
		if (WindowUtils.showOptionDialog(null, panel, MSG_RANDOMIZER, true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.INFORMATION_MESSAGE, new ImageIcon(Images.getCharacterSheetIcon(true)), new String[] { MSG_APPLY, MSG_CANCEL }, MSG_APPLY) == JOptionPane.OK_OPTION) {
			panel.applyChanges();
		}
	}
}
