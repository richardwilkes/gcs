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

package com.trollworks.gcs.menu.help;

import com.trollworks.gcs.app.AboutPanel;
import com.trollworks.gcs.app.Main;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.GraphicsUtilities;

import java.awt.Dimension;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.WindowAdapter;
import java.awt.event.WindowEvent;
import java.text.MessageFormat;

import javax.swing.JMenuItem;

/** Provides the "About" command. */
public class AboutCommand extends Command {
	private static String				MSG_ABOUT;

	static {
		LocalizedMessages.initialize(AboutCommand.class);
	}

	/** The singleton {@link AboutCommand}. */
	public static final AboutCommand	INSTANCE	= new AboutCommand();
	static AppWindow						ABOUT		= null;

	private AboutCommand() {
		super(MessageFormat.format(MSG_ABOUT, Main.getName()));
	}

	@Override public void adjustForMenu(JMenuItem item) {
		setEnabled(true);
	}

	@Override public void actionPerformed(ActionEvent event) {
		if (ABOUT != null) {
			if (ABOUT.isDisplayable() && ABOUT.isVisible()) {
				ABOUT.toFront();
				return;
			}
		}

		ABOUT = new AppWindow(getTitle(), null, null);
		ABOUT.setResizable(false);
		ABOUT.add(new AboutPanel());
		ABOUT.pack();
		Dimension size = ABOUT.getSize();
		Rectangle bounds = ABOUT.getGraphicsConfiguration().getBounds();
		ABOUT.setLocation((bounds.width - size.width) / 2, (bounds.height - size.height) / 3);
		GraphicsUtilities.forceOnScreen(ABOUT);
		ABOUT.setVisible(true);
		ABOUT.addWindowListener(new WindowAdapter() {
			@Override public void windowClosed(WindowEvent windowEvent) {
				ABOUT = null;
			}
		});
	}
}
