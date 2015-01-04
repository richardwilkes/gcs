/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.common;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.menu.edit.Undoable;
import com.trollworks.toolkit.ui.menu.file.CloseHandler;
import com.trollworks.toolkit.ui.menu.file.SaveCommand;
import com.trollworks.toolkit.ui.menu.file.Saveable;
import com.trollworks.toolkit.ui.widget.DataModifiedListener;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.dock.DockContainer;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Window;
import java.io.File;

import javax.swing.Icon;

/** Provides a common base for library and sheet files. */
public abstract class CommonDockable extends Dockable implements CloseHandler, Saveable, Undoable {
	@Localize("An error occurred while trying to save the file.")
	@Localize(locale = "de", value = "Ein Fehler ist beim Speichern der Datei aufgetreten.")
	@Localize(locale = "ru", value = "Произошла ошибка при попытке сохранить файл.")
	private static String	SAVE_ERROR;

	static {
		Localization.initialize();
	}

	private DataFile		mDataFile;
	private String			mUntitledName;

	/**
	 * Creates a new {@link CommonDockable}.
	 *
	 * @param file The {@link DataFile} to use.
	 */
	protected CommonDockable(DataFile file) {
		super(new BorderLayout());
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
	public void toFrontAndFocus() {
		Window window = UIUtilities.getAncestorOfType(this, Window.class);
		if (window != null) {
			window.toFront();
		}
		DockContainer dc = getDockContainer();
		dc.setCurrentDockable(this);
		dc.doLayout();
		dc.acquireFocus();
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
	public String getSaveTitle() {
		return getTitle();
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
		WindowUtils.showError(this, SAVE_ERROR);
		return new File[0];
	}

	@Override
	public boolean mayAttemptClose() {
		return true;
	}

	@Override
	public boolean attemptClose() {
		if (SaveCommand.attemptSave(this)) {
			getDockContainer().close(this);
			return true;
		}
		return false;
	}

	@Override
	public String getTitle() {
		File file = getBackingFile();
		String title;
		if (file == null) {
			if (mUntitledName == null) {
				mUntitledName = getDockContainer().getDock().getNextUntitledDockableName(getUntitledBaseName(), this);
			}
			title = mUntitledName;
		} else {
			title = PathUtils.getLeafName(file.getName(), false);
		}
		return title;
	}

	protected abstract String getUntitledBaseName();

	@Override
	public Icon getTitleIcon() {
		return getDataFile().getFileIcons().getImage(16);
	}

	@Override
	public String getTitleTooltip() {
		StringBuilder buffer = new StringBuilder();
		buffer.append("<html><body><b>"); //$NON-NLS-1$
		buffer.append(getTitle());
		buffer.append("</b>"); //$NON-NLS-1$
		File file = getBackingFile();
		if (file != null) {
			buffer.append("<br><font size='-2'>"); //$NON-NLS-1$
			buffer.append(file.getAbsolutePath());
			buffer.append("</font>"); //$NON-NLS-1$
		}
		buffer.append("</body></html>"); //$NON-NLS-1$
		return buffer.toString();
	}
}
