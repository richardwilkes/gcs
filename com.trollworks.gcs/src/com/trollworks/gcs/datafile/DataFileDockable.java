/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.datafile;

import com.trollworks.gcs.menu.edit.Undoable;
import com.trollworks.gcs.menu.file.CloseHandler;
import com.trollworks.gcs.menu.file.SaveCommand;
import com.trollworks.gcs.menu.file.SaveResult;
import com.trollworks.gcs.menu.file.Saveable;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.DataModifiedListener;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.ui.widget.dock.DockContainer;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Window;
import java.nio.file.Path;
import javax.swing.Icon;

/** Provides a common base for library and sheet files. */
public abstract class DataFileDockable extends Dockable implements CloseHandler, Saveable, Undoable {
    private DataFile mDataFile;
    private String   mUntitledName;

    /**
     * Creates a new DataFileDockable.
     *
     * @param file The {@link DataFile} to use.
     */
    protected DataFileDockable(DataFile file) {
        super(new BorderLayout());
        mDataFile = file;
        mDataFile.setUndoManager(new StdUndoManager());
    }

    /** @return The {@link DataFile}. */
    public DataFile getDataFile() {
        return mDataFile;
    }

    @Override
    public FileType getFileType() {
        return mDataFile.getFileType();
    }

    @Override
    public Path getBackingFile() {
        return mDataFile.getPath();
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
    public boolean saveTo(Path path) {
        if (mDataFile.save(path)) {
            mDataFile.setPath(path);
            getDockContainer().updateTitle(this);
            return true;
        }
        Modal.showError(this, I18n.text(String.format("An error occurred while trying to save the file to:\n%s", path)));
        return false;
    }

    @Override
    public boolean mayAttemptClose() {
        return true;
    }

    @Override
    public boolean attemptClose() {
        if (SaveCommand.attemptSave(this) != SaveResult.CANCEL) {
            getDockContainer().close(this);
            return true;
        }
        return false;
    }

    protected String getUntitledName() {
        if (mUntitledName == null) {
            mUntitledName = getDockContainer().getDock().getNextUntitledDockableName(getUntitledBaseName(), this);
        }
        return mUntitledName;
    }

    @Override
    public String getTitle() {
        Path path = getBackingFile();
        return path != null ? PathUtils.getLeafName(path, false) : getUntitledName();
    }

    protected abstract String getUntitledBaseName();

    @Override
    public Icon getTitleIcon() {
        return getDataFile().getFileIcon();
    }

    @Override
    public String getTitleTooltip() {
        StringBuilder buffer = new StringBuilder();
        buffer.append(getTitle());
        Path path = getBackingFile();
        if (path != null) {
            buffer.append("\n");
            buffer.append(path.normalize().toAbsolutePath());
        }
        return buffer.toString();
    }
}
