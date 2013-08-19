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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.outline.Outline;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;
import javax.swing.KeyStroke;
import javax.swing.text.JTextComponent;

/** Provides the "Delete" command. */
public class DeleteCommand extends Command {
	private static String				MSG_DELETE;

	static {
		LocalizedMessages.initialize(DeleteCommand.class);
	}

	/** The singleton {@link DeleteCommand}. */
	public static final DeleteCommand	INSTANCE	= new DeleteCommand();

	private DeleteCommand() {
		super(MSG_DELETE, KeyEvent.VK_DELETE);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		boolean isEnabled = false;
		Component comp = getFocusOwner();
		if (comp.isEnabled()) {
			if (comp instanceof JTextComponent) {
				JTextComponent textComp = (JTextComponent) comp;
				if (textComp.isEditable()) {
					isEnabled = textComp.getDocument().getLength() > 0;
				}
			} else if (comp instanceof Outline) {
				isEnabled = ((Outline) comp).canDeleteSelection();
			}
		}
		setEnabled(isEnabled);
	}

	@Override public void actionPerformed(ActionEvent event) {
		Component comp = getFocusOwner();
		if (comp instanceof JTextComponent) {
			JTextComponent textComp = (JTextComponent) comp;
			ActionListener listener = textComp.getActionForKeyStroke(KeyStroke.getKeyStroke(KeyEvent.VK_DELETE, 0));
			if (listener != null) {
				listener.actionPerformed(event);
			}
		} else if (comp instanceof Outline) {
			((Outline) comp).deleteSelection();
		}
	}
}
