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

package com.trollworks.gcs.menu.edit;

import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.OutlineProxy;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "Increment" command. */
public class IncrementCommand extends Command {
	/** The action command this command will issue. */
	public static final String				CMD_INCREMENT	= "Increment";				//$NON-NLS-1$
	private static String					MSG_INCREMENT;

	static {
		LocalizedMessages.initialize(IncrementCommand.class);
	}

	/** The singleton {@link IncrementCommand}. */
	public static final IncrementCommand	INSTANCE		= new IncrementCommand();

	private IncrementCommand() {
		super(MSG_INCREMENT, CMD_INCREMENT, KeyEvent.VK_EQUALS);
	}

	@Override
	public void adjustForMenu(JMenuItem item) {
		Component focus = getFocusOwner();
		if (focus instanceof OutlineProxy) {
			focus = ((OutlineProxy) focus).getRealOutline();
		}
		if (focus instanceof Incrementable) {
			Incrementable inc = (Incrementable) focus;
			setTitle(inc.getIncrementTitle());
			setEnabled(inc.canIncrement());
		} else {
			setTitle(MSG_INCREMENT);
			setEnabled(false);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Component focus = getFocusOwner();
		if (focus instanceof OutlineProxy) {
			focus = ((OutlineProxy) focus).getRealOutline();
		}
		((Incrementable) focus).increment();
	}
}
