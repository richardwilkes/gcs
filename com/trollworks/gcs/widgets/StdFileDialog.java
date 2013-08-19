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

package com.trollworks.gcs.widgets;

import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;

import java.awt.Component;
import java.awt.Dialog;
import java.awt.FileDialog;
import java.awt.Frame;
import java.awt.Window;
import java.io.File;
import java.io.FilenameFilter;
import java.text.MessageFormat;
import java.util.HashSet;

/** Provides standard file dialog handling. */
public class StdFileDialog implements FilenameFilter {
	private static String	MSG_UNABLE_TO_OPEN;
	private HashSet<String>	mFileNameMatchers	= new HashSet<String>();

	static {
		LocalizedMessages.initialize(StdFileDialog.class);
	}

	/**
	 * Creates a new {@link StdFileDialog}.
	 * 
	 * @param comp The {@link Component} to use for determining the parent {@link Frame} or
	 *            {@link Dialog}.
	 * @param open Whether an 'open' or a 'save' dialog is presented.
	 * @param title The title to use.
	 * @param dir The initial directory to start with. May be <code>null</code>.
	 * @param name The initial file name to start with. May be <code>null</code>.
	 * @param extension One or more file name extensions that should be allowed to be opened or
	 *            saved. If this is a save dialog, the first extension will be forced if none match
	 *            what the user has supplied.
	 * @return The chosen {@link File} or <code>null</code>.
	 */
	public static File choose(Component comp, boolean open, String title, String dir, String name, String... extension) {
		StdFileDialog filter = new StdFileDialog(extension);
		FileDialog dialog;
		Window window = WindowUtils.getWindowForComponent(comp);
		int mode = open ? FileDialog.LOAD : FileDialog.SAVE;
		if (window instanceof Frame) {
			dialog = new FileDialog((Frame) window, title, mode);
		} else {
			dialog = new FileDialog((Dialog) window, title, mode);
		}
		dialog.setFilenameFilter(filter);
		dialog.setDirectory(dir);
		dialog.setFile(name);
		dialog.setVisible(true);
		String result = dialog.getFile();
		if (result != null) {
			if (filter.accept(null, result)) {
				return new File(dialog.getDirectory(), result);
			} else if (!open) {
				return new File(dialog.getDirectory(), Path.enforceExtension(result, extension[0]));
			}
			showCannotOpenMsg(comp, result);
		}
		return null;
	}

	/**
	 * @param comp The {@link Component} to use for determining the parent {@link Frame} or
	 *            {@link Dialog}.
	 * @param name The name of the file that cannot be opened.
	 */
	public static void showCannotOpenMsg(Component comp, String name) {
		WindowUtils.showError(comp, MessageFormat.format(MSG_UNABLE_TO_OPEN, name));
	}

	/**
	 * Convenience for creating a regular expression that will match a file extension. This takes
	 * care of turning on case-insensitivity for those platforms that need it.
	 * 
	 * @param extension A file name extension.
	 * @return The regular expression that will match the specified file name extension.
	 */
	public static String createExtensionMatcher(String extension) {
		StringBuilder builder = new StringBuilder();

		if (Platform.isMacintosh() || Platform.isWindows()) {
			builder.append("(?i)"); //$NON-NLS-1$
		}
		builder.append("^.*\\"); //$NON-NLS-1$
		if (!extension.startsWith(".")) { //$NON-NLS-1$
			builder.append('.');
		}
		builder.append(extension);
		builder.append('$');
		return builder.toString();
	}

	private StdFileDialog(String... extension) {
		for (String one : extension) {
			mFileNameMatchers.add(createExtensionMatcher(one));
		}
	}

	public boolean accept(File dir, String name) {
		if (mFileNameMatchers.isEmpty()) {
			return true;
		}
		for (String one : mFileNameMatchers) {
			if (name.matches(one)) {
				return true;
			}
		}
		return false;
	}
}
