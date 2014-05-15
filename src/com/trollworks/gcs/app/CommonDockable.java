/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.common.DataFile;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.menu.edit.Undoable;
import com.trollworks.toolkit.ui.menu.file.SaveCommand;
import com.trollworks.toolkit.ui.menu.file.Saveable;
import com.trollworks.toolkit.ui.widget.DataModifiedListener;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.dock.DockCloseable;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.undo.StdUndoManager;

import java.io.File;

import javax.swing.Icon;
import javax.swing.JOptionPane;

/** Provides a common base for library and sheet files. */
public abstract class CommonDockable implements Dockable, DockCloseable, Saveable, Undoable {
	@Localize("Save")
	private static String	SAVE;
	@Localize("Save changes to \"%s\"?")
	private static String	SAVE_CHANGES;
	@Localize("An error occurred while trying to save the file.")
	private static String	SAVE_ERROR;

	static {
		Localization.initialize();
	}

	private DataFile		mDataFile;

	/**
	 * Creates a new {@link CommonDockable}.
	 *
	 * @param file The {@link DataFile} to use.
	 */
	protected CommonDockable(DataFile file) {
		mDataFile = file;
		mDataFile.setUndoManager(new StdUndoManager());
	}

	/** @return The {@link DataFile}. */
	public DataFile getDataFile() {
		return mDataFile;
	}

	@Override
	public File getBackingFile() {
		return mDataFile.getFile();
	}

	@Override
	public StdUndoManager getUndoManager() {
		return mDataFile.getUndoManager();
	}

	@Override
	public boolean isModified() {
		return mDataFile.isModified();
	}

	@Override
	public void addDataModifiedListener(DataModifiedListener listener) {
		mDataFile.addDataModifiedListener(listener);
	}

	@Override
	public void removeDataModifiedListener(DataModifiedListener listener) {
		mDataFile.removeDataModifiedListener(listener);
	}

	@Override
	public String getPreferredSavePath() {
		return PathUtils.getFullPath(getBackingFile());
	}

	@Override
	public File[] saveTo(File file) {
		if (mDataFile.save(file)) {
			mDataFile.setFile(file);
			getDockContainer().updateTitle(this);
			return new File[] { file };
		}
		WindowUtils.showError(getContent(), SAVE_ERROR);
		return new File[0];
	}

	@Override
	public boolean attemptClose() {
		UIUtilities.forceFocusToAccept();
		if (mDataFile.isModified()) {
			switch (JOptionPane.showConfirmDialog(getContent(), String.format(SAVE_CHANGES, getTitle()), SAVE, JOptionPane.YES_NO_CANCEL_OPTION)) {
				case JOptionPane.CANCEL_OPTION:
				case JOptionPane.CLOSED_OPTION:
					return false;
				case JOptionPane.YES_OPTION:
					SaveCommand.save(this);
					return !mDataFile.isModified();
				default:
					return true;
			}
		}
		return true;
	}

	@Override
	public String getTitle() {
		return PathUtils.getLeafName(getBackingFile().getName(), false);
	}

	@Override
	public Icon getTitleIcon() {
		return getDataFile().getFileIcons().getIcon(16);
	}

	@Override
	public String getTitleTooltip() {
		StringBuilder buffer = new StringBuilder();
		buffer.append("<html><body><b>"); //$NON-NLS-1$
		buffer.append(getTitle());
		buffer.append("</b><br><font size='-2'>"); //$NON-NLS-1$
		buffer.append(getBackingFile().getAbsolutePath());
		buffer.append("</font></body></html>"); //$NON-NLS-1$
		return buffer.toString();
	}
}
