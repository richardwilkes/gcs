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

import static com.trollworks.gcs.menu.file.NewCharacterTemplateCommand_LS.*;

import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.menu.Command;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

@Localized({
				@LS(key = "NEW_CHARACTER_TEMPLATE", msg = "New Character Template"),
})
/** Provides the "New Character Template" command. */
public class NewCharacterTemplateCommand extends Command {
	/** The action command this command will issue. */
	public static final String						CMD_NEW_CHARACTER_TEMPLATE	= "NewCharacterTemplate";				//$NON-NLS-1$

	/** The singletone {@link NewCharacterTemplateCommand}. */
	public static final NewCharacterTemplateCommand	INSTANCE					= new NewCharacterTemplateCommand();

	private NewCharacterTemplateCommand() {
		super(NEW_CHARACTER_TEMPLATE, CMD_NEW_CHARACTER_TEMPLATE, KeyEvent.VK_N, SHIFTED_COMMAND_MODIFIER);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		// Do nothing. We're always enabled.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		newTemplate();
	}

	/** @return The newly created a new {@link TemplateWindow}. */
	public static TemplateWindow newTemplate() {
		return TemplateWindow.displayTemplateWindow(new Template());
	}
}
