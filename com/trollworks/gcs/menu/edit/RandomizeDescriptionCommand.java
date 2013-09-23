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

import static com.trollworks.gcs.menu.edit.RandomizeDescriptionCommand_LS.*;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.character.DescriptionRandomizer;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.widgets.WindowUtils;

import java.awt.Window;
import java.awt.event.ActionEvent;

import javax.swing.ImageIcon;
import javax.swing.JMenuItem;
import javax.swing.JOptionPane;

@Localized({
				@LS(key = "RANDOMIZE_DESCRIPTION", msg = "Randomize Description\u2026"),
				@LS(key = "RANDOMIZER", msg = "Description Randomizer"),
				@LS(key = "APPLY", msg = "Apply"),
				@LS(key = "CANCEL", msg = "Cancel"),
})
/** Provides the "Randomize Description" command. */
public class RandomizeDescriptionCommand extends Command {
	/** The action command this command will issue. */
	public static final String						CMD_RANDOMIZE_DESCRIPTION	= "RandomizeDescription";				//$NON-NLS-1$

	/** The singleton {@link RandomizeDescriptionCommand}. */
	public static final RandomizeDescriptionCommand	INSTANCE					= new RandomizeDescriptionCommand();

	private RandomizeDescriptionCommand() {
		super(RANDOMIZE_DESCRIPTION, CMD_RANDOMIZE_DESCRIPTION);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		setEnabled(window instanceof SheetWindow);
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		DescriptionRandomizer panel = new DescriptionRandomizer(((SheetWindow) getActiveWindow()).getCharacter());
		if (WindowUtils.showOptionDialog(null, panel, RANDOMIZER, true, JOptionPane.OK_CANCEL_OPTION, JOptionPane.INFORMATION_MESSAGE, new ImageIcon(GCSImages.getCharacterSheetIcon(true)), new String[] { APPLY, CANCEL }, APPLY) == JOptionPane.OK_OPTION) {
			panel.applyChanges();
		}
	}
}
