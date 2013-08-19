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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.library.LibraryWindow;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.utility.LocalizedMessages;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "New Advantage" command. */
public class NewAdvantageCommand extends Command {
	/** The action command this command will issue. */
	public static final String				CMD_NEW_ADVANTAGE			= "NewAdvantage";																												//$NON-NLS-1$
	/** The action command this command will issue. */
	public static final String				CMD_NEW_ADVANTAGE_CONTAINER	= "NewAdvantageContainer";																										//$NON-NLS-1$
	private static String					MSG_ADVANTAGE;
	private static String					MSG_ADVANTAGE_CONTAINER;

	static {
		LocalizedMessages.initialize(NewAdvantageCommand.class);
	}

	/** The "New Advantage" command. */
	public static final NewAdvantageCommand	INSTANCE					= new NewAdvantageCommand(false, MSG_ADVANTAGE, CMD_NEW_ADVANTAGE, KeyEvent.VK_D, COMMAND_MODIFIER);
	/** The "New Advantage Container" command. */
	public static final NewAdvantageCommand	CONTAINER_INSTANCE			= new NewAdvantageCommand(true, MSG_ADVANTAGE_CONTAINER, CMD_NEW_ADVANTAGE_CONTAINER, KeyEvent.VK_D, SHIFTED_COMMAND_MODIFIER);
	private boolean							mContainer;

	private NewAdvantageCommand(boolean container, String title, String cmd, int keyCode, int modifiers) {
		super(title, cmd, keyCode, modifiers);
		mContainer = container;
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			setEnabled(!((LibraryWindow) window).getOutline().getModel().isLocked());
		} else {
			setEnabled(window instanceof SheetWindow || window instanceof TemplateWindow);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		ListOutline outline;
		DataFile dataFile;

		Window window = getActiveWindow();
		if (window instanceof LibraryWindow) {
			LibraryWindow libraryWindow = (LibraryWindow) window;
			libraryWindow.switchToAdvantages();
			dataFile = libraryWindow.getLibraryFile();
			outline = libraryWindow.getOutline();
		} else if (window instanceof SheetWindow) {
			SheetWindow sheetWindow = (SheetWindow) window;
			outline = sheetWindow.getSheet().getAdvantageOutline();
			dataFile = sheetWindow.getCharacter();
		} else if (window instanceof TemplateWindow) {
			TemplateWindow templateWindow = (TemplateWindow) window;
			outline = templateWindow.getSheet().getAdvantageOutline();
			dataFile = templateWindow.getTemplate();
		} else {
			return;
		}

		Advantage advantage = new Advantage(dataFile, mContainer);
		outline.addRow(advantage, getTitle(), false);
		outline.getModel().select(advantage, false);
		outline.scrollSelectionIntoView();
		outline.openDetailEditor(true);
	}
}
