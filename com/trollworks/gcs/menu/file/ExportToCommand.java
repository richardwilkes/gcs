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

import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.widgets.StdFileDialog;

import java.awt.Component;
import java.awt.Window;
import java.awt.event.ActionEvent;
import java.io.File;

import javax.swing.JMenuItem;

/** Provides the "Save As..." command. */
public class ExportToCommand extends Command {
	private static String				MSG_EXPORT_TO_HTML;
	private static String				MSG_EXPORT_TO_PDF;
	private static String				MSG_EXPORT_TO_PNG;

	static {
		LocalizedMessages.initialize(ExportToCommand.class);
	}

	/** The "Export To HTML...". */
	public static final ExportToCommand	EXPORT_TO_HTML	= new ExportToCommand(MSG_EXPORT_TO_HTML, SheetWindow.HTML_EXTENSION);
	/** The "Export To PDF...". */
	public static final ExportToCommand	EXPORT_TO_PDF	= new ExportToCommand(MSG_EXPORT_TO_PDF, SheetWindow.PDF_EXTENSION);
	/** The "Export To PNG...". */
	public static final ExportToCommand	EXPORT_TO_PNG	= new ExportToCommand(MSG_EXPORT_TO_PNG, SheetWindow.PNG_EXTENSION);
	private String						mExtension;

	private ExportToCommand(String title, String extension) {
		super(title);
		mExtension = extension;
	}

	@Override public void adjustForMenu(JMenuItem item) {
		Window activeWindow = getActiveWindow();
		boolean enable = false;
		if (activeWindow instanceof Saveable) {
			for (String extension : ((Saveable) activeWindow).getAllowedExtensions()) {
				if (mExtension.equals(extension)) {
					enable = true;
					break;
				}
			}
		}
		setEnabled(enable);
	}

	@Override public void actionPerformed(ActionEvent event) {
		saveAs((Saveable) getActiveWindow());
	}

	/**
	 * Allows the user to save the file under another name.
	 * 
	 * @param saveable The {@link Saveable} to work on.
	 * @return The file(s) actually written to.
	 */
	public File[] saveAs(Saveable saveable) {
		String path = saveable.getPreferredSavePath();
		File result = StdFileDialog.choose(saveable instanceof Component ? (Component) saveable : null, false, (String) getValue(NAME), Path.getParent(path), Path.getLeafName(path), mExtension);
		return result != null ? saveable.saveTo(result) : new File[0];
	}
}
