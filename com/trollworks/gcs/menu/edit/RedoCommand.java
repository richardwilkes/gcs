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

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;
import javax.swing.undo.UndoManager;

/** Provides the "Redo" command. */
public class RedoCommand extends Command {
	private static String			MSG_REDO;

	static {
		LocalizedMessages.initialize(RedoCommand.class);
	}

	/** The singleton {@link RedoCommand}. */
	public static final RedoCommand	INSTANCE	= new RedoCommand();

	private RedoCommand() {
		super(MSG_REDO, KeyEvent.VK_Y);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof Undoable) {
			UndoManager mgr = ((Undoable) window).getUndoManager();
			setEnabled(mgr.canRedo());
			setTitle(mgr.getRedoPresentationName());
		} else {
			setEnabled(false);
			setTitle(MSG_REDO);
		}
	}

	@Override public void actionPerformed(ActionEvent event) {
		((Undoable) getActiveWindow()).getUndoManager().redo();
	}
}
