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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.CommitEnforcer;

import java.awt.Frame;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.text.MessageFormat;

import javax.swing.JMenuItem;
import javax.swing.JOptionPane;

/** Provides the "Close" command. */
public class CloseCommand extends Command {
	private static String				MSG_CLOSE;
	private static String				MSG_SAVE;
	private static String				MSG_SAVE_CHANGES;

	static {
		LocalizedMessages.initialize(CloseCommand.class);
	}

	/** The singleton {@link CloseCommand}. */
	public static final CloseCommand	INSTANCE	= new CloseCommand();

	private CloseCommand() {
		super(MSG_CLOSE, KeyEvent.VK_W);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		// Do nothing. Always enabled.
	}

	@Override public void actionPerformed(ActionEvent event) {
		close(getActiveWindow(), true);
	}

	/**
	 * @param window The {@link Window} to close.
	 * @param quitIfLast Call {@link QuitCommand#quit()} if no windows are open when this method
	 *            completes.
	 * @return <code>true</code> if the {@link Window} was closed.
	 */
	public boolean close(Window window, boolean quitIfLast) {
		if (window != null && !AppWindow.hasOwnedWindowsShowing(window)) {
			if (window instanceof Frame && window instanceof Saveable) {
				CommitEnforcer.forceFocusToAccept();
				Saveable saveable = (Saveable) window;
				if (saveable.isModified()) {
					int answer = JOptionPane.showConfirmDialog(window, MessageFormat.format(MSG_SAVE_CHANGES, ((Frame) window).getTitle()), MSG_SAVE, JOptionPane.YES_NO_CANCEL_OPTION);
					if (answer == JOptionPane.CANCEL_OPTION || answer == JOptionPane.CLOSED_OPTION) {
						return false;
					}
					if (answer == JOptionPane.YES_OPTION) {
						SaveCommand.INSTANCE.save(saveable);
						if (saveable.isModified()) {
							return false;
						}
					}
				}
			}
			window.dispose();
		}
		if (quitIfLast) {
			for (Frame frame : Frame.getFrames()) {
				if (frame.isVisible() || AppWindow.hasOwnedWindowsShowing(frame)) {
					return true;
				}
			}
			QuitCommand.INSTANCE.adjustForMenu(null);
			QuitCommand.INSTANCE.quit();
		}
		return true;
	}
}
