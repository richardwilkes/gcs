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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.OutlineModel;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "Copy To Character Sheet" command. */
public class CopyToSheetCommand extends Command {
	/** The action command this command will issue. */
	public static final String				CMD_COPY_TO_SHEET	= "CopyToSheet";			//$NON-NLS-1$
	private static String					MSG_COPY_TO_SHEET;

	static {
		LocalizedMessages.initialize(CopyToSheetCommand.class);
	}

	/** The singleton {@link CopyToSheetCommand}. */
	public static final CopyToSheetCommand	INSTANCE			= new CopyToSheetCommand();

	private CopyToSheetCommand() {
		super(MSG_COPY_TO_SHEET, CMD_COPY_TO_SHEET, KeyEvent.VK_C, SHIFTED_COMMAND_MODIFIER);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			setEnabled(((LibraryWindow) window).getOutline().getModel().hasSelection() && SheetWindow.getTopSheet() != null);
		} else {
			setEnabled(false);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			OutlineModel outlineModel = ((LibraryWindow) window).getOutline().getModel();
			if (outlineModel.hasSelection()) {
				SheetWindow sheet = SheetWindow.getTopSheet();
				if (sheet != null) {
					sheet.addRows(outlineModel.getSelectionAsList(true));
				}
			}
		}
	}
}
