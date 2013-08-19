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

import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.Outline;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "Open Detail Editor" command. */
public class OpenEditorCommand extends Command {
	/** The action command this command will issue. */
	public static final String				CMD_OPEN_EDITOR	= "OpeNEditor";			//$NON-NLS-1$
	private static String					MSG_OPEN_EDITOR;

	static {
		LocalizedMessages.initialize(OpenEditorCommand.class);
	}

	/** The singleton {@link OpenEditorCommand}. */
	public static final OpenEditorCommand	INSTANCE		= new OpenEditorCommand();

	private OpenEditorCommand() {
		super(MSG_OPEN_EDITOR, CMD_OPEN_EDITOR, KeyEvent.VK_I);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Component comp = getFocusOwner();
		if (comp instanceof Outline) {
			setEnabled(((Outline) comp).getModel().hasSelection());
		} else {
			setEnabled(false);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Outline outline = (Outline) getFocusOwner();
		((ListOutline) outline.getRealOutline()).openDetailEditor(false);
	}
}
