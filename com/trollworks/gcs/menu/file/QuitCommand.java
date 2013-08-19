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
import com.trollworks.gcs.utility.Fonts;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.widgets.AppWindow;

import java.awt.Frame;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

import javax.swing.JMenuItem;

/** Provides the "Quit"/"Exit" command. */
public class QuitCommand extends Command {
	private static String			MSG_QUIT;
	private static String			MSG_EXIT;

	static {
		LocalizedMessages.initialize(QuitCommand.class);
	}

	/** The singleton {@link QuitCommand}. */
	public static final QuitCommand	INSTANCE	= new QuitCommand();

	private QuitCommand() {
		super(Platform.isMacintosh() ? MSG_QUIT : MSG_EXIT, KeyEvent.VK_Q);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		for (Frame frame : Frame.getFrames()) {
			if (frame.isVisible() && AppWindow.hasOwnedWindowsShowing(frame)) {
				setEnabled(false);
				return;
			}
		}
		setEnabled(true);
	}

	@Override public void actionPerformed(ActionEvent event) {
		quit();
	}

	/** Attempts to quit. */
	public void quit() {
		if (isEnabled()) {
			for (Frame frame : Frame.getFrames()) {
				if (frame.isVisible()) {
					try {
						if (!CloseCommand.INSTANCE.close(frame, false)) {
							return;
						}
					} catch (Exception exception) {
						exception.printStackTrace(System.err);
					}
				}
			}

			try {
				Fonts.saveToPreferences();
				Preferences.getInstance().save();
			} catch (Exception exception) {
				// Ignore, since preferences may not have been initialized...
			}
			System.exit(0);
		}
	}
}
