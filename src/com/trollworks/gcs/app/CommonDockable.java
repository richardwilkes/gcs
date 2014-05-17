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
import com.trollworks.toolkit.ui.menu.file.CloseHandler;
import com.trollworks.toolkit.ui.menu.file.SaveCommand;
import com.trollworks.toolkit.ui.menu.file.Saveable;
import com.trollworks.toolkit.ui.widget.DataModifiedListener;
import com.trollworks.toolkit.ui.widget.WindowUtils;
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
		super(new BorderLayout());
		mDataFile = file;
		mDataFile.setUndoManager(new StdUndoManager());
	}

	/** @return The {@link DataFile}. */
	public DataFile getDataFile() {
		return mDataFile;
	}

	@Override
	public File getCurrentBackingFile() {
		return mDataFile.getFile();
	}

	@Override
	public void toFrontAndFocus() {
		Window window = UIUtilities.getAncestorOfType(this, Window.class);
		if (window != null) {
			window.toFront();
		}
		getDockContainer().acquireFocus();
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
		return PathUtils.getFullPath(getCurrentBackingFile());
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
		return PathUtils.getLeafName(getCurrentBackingFile().getName(), false);
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
		buffer.append(getCurrentBackingFile().getAbsolutePath());
		buffer.append("</font></body></html>"); //$NON-NLS-1$
		return buffer.toString();
	}
}
