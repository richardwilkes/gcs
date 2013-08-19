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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.print.PrintManager;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.WindowUtils;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.awt.print.Printable;

import javax.swing.JMenuItem;

/** Provides the "Print..." command. */
public class PrintCommand extends Command {
	private static String				MSG_PRINT;
	private static String				MSG_NO_PRINTER_SELECTED;

	static {
		LocalizedMessages.initialize(PrintCommand.class);
	}

	/** The singleton {@link PrintCommand}. */
	public static final PrintCommand	INSTANCE	= new PrintCommand();

	private PrintCommand() {
		super(MSG_PRINT, KeyEvent.VK_P);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		setEnabled(window instanceof AppWindow && window instanceof Printable);
	}

	@Override public void actionPerformed(ActionEvent event) {
		print((AppWindow) getActiveWindow());
	}

	/** @param window The {@link AppWindow} to print. */
	public void print(AppWindow window) {
		if (window instanceof Printable) {
			PrintManager mgr = window.getPrintManager();
			if (mgr != null) {
				mgr.print(window, window.getTitle(), (Printable) window);
			} else {
				WindowUtils.showError(window, MSG_NO_PRINTER_SELECTED);
			}
		}
	}
}
