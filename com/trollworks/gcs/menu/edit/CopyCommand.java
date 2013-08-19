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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;
import javax.swing.text.JTextComponent;

/** Provides the "Copy" command. */
public class CopyCommand extends Command {
	private static String			MSG_COPY;

	static {
		LocalizedMessages.initialize(CopyCommand.class);
	}

	/** The singleton {@link CopyCommand}. */
	public static final CopyCommand	INSTANCE	= new CopyCommand();

	private CopyCommand() {
		super(MSG_COPY, KeyEvent.VK_C);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		boolean isEnabled = false;
		Component comp = getFocusOwner();
		if (comp instanceof JTextComponent) {
			JTextComponent textComp = (JTextComponent) comp;
			String text = textComp.getSelectedText();
			isEnabled = text != null && text.length() > 0;
		}
		setEnabled(isEnabled);
	}

	@Override public void actionPerformed(ActionEvent event) {
		((JTextComponent) getFocusOwner()).copy();
	}
}
